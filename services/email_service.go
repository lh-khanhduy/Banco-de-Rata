package services

import (
	"context"
	"log"

	db "github.com/lh-khanhduy/banco_de_rata/db/sqlc"
	"github.com/lh-khanhduy/banco_de_rata/pb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Server) VerifyEmail(ctx context.Context, req *pb.VerifyEmailRequest) (*pb.VerifyEmailResponse, error) {
	log.Println("req: ", req)

	violations := validateVerifyEmailRequest(req)
	if violations != nil {
		return nil, invalidArgsError(violations)
	}

	result, err := s.store.VerifyEmailTx(ctx, db.VerifyEmailTxParams{
		EmailId:    req.GetEmailId(),
		SecretCode: req.GetSecretCode(),
	})
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to verify email")
	}

	res := &pb.VerifyEmailResponse{
		IsVerified: result.User.IsEmailVerified,
	}

	return res, nil
}
