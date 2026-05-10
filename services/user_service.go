package services

import (
	"context"
	"database/sql"

	db "github.com/lh-khanhduy/banco_de_rata/db/sqlc"
	"github.com/lh-khanhduy/banco_de_rata/pb"
	"github.com/lh-khanhduy/banco_de_rata/utils"
	"github.com/lib/pq"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *Server) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.CreateUserResponse, error) {

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

	// if md, ok := metadata.FromIncomingContext(ctx); ok {
	// 	if ua := md.Get("user-agent"); len(ua) > 0 {
	// 		userAgent = ua[0]
	// 	}

	// 	if forwardedFor := md.Get("x-forwarded-for"); len(forwardedFor) > 0 {
	// 		clientIP = strings.TrimSpace(strings.Split(forwardedFor[0], ",")[0])
	// 	}
	// }

	session, err := s.store.CreateSession(ctx, db.CreateSessionParams{
		ID:           refreshPayload.ID,
		Username:     user.Username,
		RefreshToken: refreshToken,
		UserAgent:    "",
		ClientIp:     "",
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
