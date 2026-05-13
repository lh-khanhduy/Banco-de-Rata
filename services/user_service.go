package services

import (
	"context"
	"database/sql"
	"time"

	db "github.com/lh-khanhduy/banco_de_rata/db/sqlc"
	"github.com/lh-khanhduy/banco_de_rata/pb"
	"github.com/lh-khanhduy/banco_de_rata/utils"
	"github.com/lib/pq"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *Server) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.CreateUserResponse, error) {
	violations := validateCreateUserRequest(req)
	if violations != nil {
		return nil, invalidArgsError(violations)
	}

	hashedPassword, err := utils.HashPassword(req.GetPassword())
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to hash password")
	}

	args := db.CreateUserParams{
		Username:       req.GetUsername(),
		HashedPassword: hashedPassword,
		FullName:       req.GetFullName(),
		Email:          req.GetEmail(),
	}

	user, err := s.store.CreateUser(ctx, args)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Code.Name() {
			case "unique_violation":
				return nil, status.Error(codes.AlreadyExists, "username already exists")
			}
		}

		return nil, status.Error(codes.Internal, "failed to created user")
	}

	res := &pb.CreateUserResponse{
		User: &pb.User{
			Username:          user.Username,
			FullName:          user.FullName,
			Email:             user.Email,
			PasswordChangedAt: timestamppb.New(user.PasswordChangedAt),
			CreatedAt:         timestamppb.New(user.CreatedAt),
		},
	}

	return res, nil
}

func (s *Server) LoginUser(ctx context.Context, req *pb.LoginUserRequest) (*pb.LoginUserResponse, error) {
	violations := validateLoginUserRequest(req)
	if violations != nil {
		return nil, invalidArgsError(violations)
	}

	user, err := s.store.GetUser(ctx, req.GetUsername())
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Error(codes.NotFound, "username not found")
		}
		return nil, status.Error(codes.Internal, "failed to find user")
	}

	err = utils.CheckPassword(req.GetPassword(), user.HashedPassword)
	if err != nil {
		return nil, status.Error(codes.NotFound, "incorrect password")
	}

	// ACCESS TOKEN
	accessToken, accessPayload, err := s.tokenMaker.CreateToken(req.GetUsername(), s.config.AccessTokenDuration)
	if err != nil {
		return nil, status.Error(codes.Internal, "cannot created access token")
	}

	// REFRESH TOKEN
	refreshToken, refreshPayload, err := s.tokenMaker.CreateToken(req.GetUsername(), s.config.RefreshTokenDuration)
	if err != nil {
		return nil, status.Error(codes.Internal, "cannot created refresh token")
	}

	md := s.extractMetadata(ctx)

	session, err := s.store.CreateSession(ctx, db.CreateSessionParams{
		ID:           refreshPayload.ID,
		Username:     user.Username,
		RefreshToken: refreshToken,
		UserAgent:    md.UserAgent,
		ClientIp:     md.ClientIP,
		IsBlocked:    false,
		ExpiresAt:    refreshPayload.ExpiredAt,
	})
	if err != nil {
		return nil, status.Error(codes.Internal, "cannot created sessions")
	}

	res := &pb.LoginUserResponse{
		SessionId:             session.ID.String(),
		AccessToken:           accessToken,
		AccessTokenExpiredAt:  timestamppb.New(accessPayload.ExpiredAt),
		RefreshToken:          refreshToken,
		RefreshTokenExpiredAt: timestamppb.New(refreshPayload.ExpiredAt),
		User: &pb.User{
			Username:          user.Username,
			FullName:          user.FullName,
			Email:             user.Email,
			PasswordChangedAt: timestamppb.New(user.PasswordChangedAt),
			CreatedAt:         timestamppb.New(user.CreatedAt),
		},
	}

	return res, nil
}

func (s *Server) RenewToken(ctx context.Context, req *pb.RenewTokenRequest) (*pb.RenewTokenResponse, error) {
	refreshPayload, err := s.tokenMaker.VerifyToken(req.GetRefreshToken())
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "invalid access token")
	}

	session, err := s.store.GetSession(ctx, refreshPayload.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Error(codes.NotFound, "session not found")
		}
		return nil, status.Error(codes.Internal, "failed to find session")
	}

	if session.IsBlocked {
		return nil, status.Error(codes.Unauthenticated, "session is blocked")
	}

	if session.Username != refreshPayload.Username {
		return nil, status.Error(codes.Unauthenticated, "incorrect session user")
	}

	if session.RefreshToken != req.RefreshToken {
		return nil, status.Error(codes.Unauthenticated, "mismatched token session")
	}

	if time.Now().After(session.ExpiresAt) {
		return nil, status.Error(codes.Unauthenticated, "session expired")
	}

	// NEW ACCESS TOKEN
	accessToken, accessPayload, err := s.tokenMaker.CreateToken(refreshPayload.Username, s.config.AccessTokenDuration)
	if err != nil {
		return nil, status.Error(codes.Internal, "cannot created new access token")
	}

	res := &pb.RenewTokenResponse{
		AccessToken:          accessToken,
		AccessTokenExpiredAt: timestamppb.New(accessPayload.ExpiredAt),
	}

	return res, nil
}

func (s *Server) UpdateUser(ctx context.Context, req *pb.UpdateUserRequest) (*pb.UpdateUserResponse, error) {
	authPayload, err := s.authorizeUser(ctx)
	if err != nil {
		return nil, unauthenticatedError(err)
	}

	violations := validateUpdateUserRequest(req)
	if violations != nil {
		return nil, invalidArgsError(violations)
	}

	if authPayload.Username != req.GetUsername() {
		return nil, status.Error(codes.PermissionDenied, "cannot update other user's info")
	}

	args := db.UpdateUserParams{
		Username: req.GetUsername(),
		FullName: sql.NullString{
			String: req.GetFullName(),
			Valid:  req.FullName != nil,
		},
		Email: sql.NullString{
			String: req.GetEmail(),
			Valid:  req.Email != nil,
		},
	}

	if req.Password != nil {
		hashedPassword, err := utils.HashPassword(req.GetPassword())
		if err != nil {
			return nil, status.Error(codes.Internal, "failed to hash password")
		}

		args.HashedPassword = sql.NullString{
			String: hashedPassword,
			Valid:  true,
		}

		args.PasswordChangedAt = sql.NullTime{
			Time:  time.Now(),
			Valid: true,
		}
	}

	user, err := s.store.UpdateUser(ctx, args)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Error(codes.NotFound, "user not found")
		}

		return nil, status.Error(codes.Internal, "failed to update user")
	}

	res := &pb.UpdateUserResponse{
		User: &pb.User{
			Username:          user.Username,
			FullName:          user.FullName,
			Email:             user.Email,
			PasswordChangedAt: timestamppb.New(user.PasswordChangedAt),
			CreatedAt:         timestamppb.New(user.CreatedAt),
		},
	}

	return res, nil
}
