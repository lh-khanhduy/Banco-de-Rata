package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	db "github.com/lh-khanhduy/banco_de_rata/db/sqlc"
	"github.com/lh-khanhduy/banco_de_rata/mail"
	"github.com/lh-khanhduy/banco_de_rata/pb"
	"github.com/lh-khanhduy/banco_de_rata/utils"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// CREATE USER NEED A TX
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
		if db.ErrorCode(err) == db.UniqueViolation {
			return nil, status.Error(codes.AlreadyExists, err.Error())
		}

		return nil, status.Error(codes.Internal, "failed to created user")
	}

	// -------- CREATE, SEND THE EMAIL WITH ID AND CODE

	verifyEmail, err := s.store.CreateVerifyEmail(ctx, db.CreateVerifyEmailParams{
		Username:   user.Username,
		Email:      user.Email,
		SecretCode: utils.RandomString(32),
	})
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to create verify email")
	}

	mailer, err := mail.NewMailerFromConfig()
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to create mail worker")
	}

	to := []string{"khanhduywin1907@gmail.com"}
	content := fmt.Sprintln("ID is:", verifyEmail.ID, " and code is:", verifyEmail.SecretCode)

	err = mailer.SendEmail(to, "VERIFICATION EMAIL", content, "", []string{})
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to send verification email")
	}

	// --------

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
		if errors.Is(err, db.ErrRecordNotFound) {
			return nil, status.Error(codes.NotFound, "username not found")
		}
		return nil, status.Error(codes.Internal, "failed to find user")
	}

	err = utils.CheckPassword(req.GetPassword(), user.HashedPassword)
	if err != nil {
		return nil, status.Error(codes.NotFound, "incorrect password")
	}

	// ACCESS TOKEN
	accessToken, accessPayload, err := s.tokenMaker.CreateToken(user.Username, user.Role, s.config.AccessTokenDuration)
	if err != nil {
		return nil, status.Error(codes.Internal, "cannot created access token")
	}

	// REFRESH TOKEN
	refreshToken, refreshPayload, err := s.tokenMaker.CreateToken(user.Username, user.Role, s.config.RefreshTokenDuration)
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
		if errors.Is(err, db.ErrRecordNotFound) {
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
	accessToken, accessPayload, err := s.tokenMaker.CreateToken(refreshPayload.Username, refreshPayload.Role, s.config.AccessTokenDuration)
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
	authPayload, err := s.authorizeUser(ctx, []string{utils.RoleBankWorker, utils.RoleDepositor})
	if err != nil {
		return nil, unauthenticatedError(err)
	}

	violations := validateUpdateUserRequest(req)
	if violations != nil {
		return nil, invalidArgsError(violations)
	}

	if authPayload.Role != utils.RoleBankWorker && authPayload.Username != req.GetUsername() {
		return nil, status.Error(codes.PermissionDenied, "cannot update other user's info")
	}

	args := db.UpdateUserParams{
		Username: req.GetUsername(),
		FullName: pgtype.Text{
			String: req.GetFullName(),
			Valid:  req.FullName != nil,
		},
		Email: pgtype.Text{
			String: req.GetEmail(),
			Valid:  req.Email != nil,
		},
	}

	if req.Password != nil {
		hashedPassword, err := utils.HashPassword(req.GetPassword())
		if err != nil {
			return nil, status.Error(codes.Internal, "failed to hash password")
		}

		args.HashedPassword = pgtype.Text{
			String: hashedPassword,
			Valid:  true,
		}

		args.PasswordChangedAt = pgtype.Timestamptz{
			Time:  time.Now(),
			Valid: true,
		}
	}

	user, err := s.store.UpdateUser(ctx, args)
	if err != nil {
		if errors.Is(err, db.ErrRecordNotFound) {
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
