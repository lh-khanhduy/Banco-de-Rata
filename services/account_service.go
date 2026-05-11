package services

import (
	"context"
	"fmt"

	db "github.com/lh-khanhduy/banco_de_rata/db/sqlc"
	"github.com/lh-khanhduy/banco_de_rata/pb"
	"github.com/lib/pq"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Server) CreateAccount(ctx context.Context, req *pb.CreateAccountRequest) (*pb.IdResponse, error) {
	md := s.extractMetadata(ctx)

	if md.Username == "" {
		return nil, status.Errorf(codes.Unauthenticated, "missing or invalid username")
	}

	args := db.CreateAccountParams{
		Owner:    md.Username,
		Currency: req.GetCurrency(),
		Balance:  0,
	}

	account, err := s.store.CreateAccount(ctx, args)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Code.Name() {
			case "foreign_key_violation", "unique_violation":
				return nil, status.Error(codes.AlreadyExists, "violate database constraint")
			}
		}

		return nil, status.Error(codes.Internal, "cannot create account")
	}

	res := &pb.IdResponse{
		Id: fmt.Sprint(account.ID),
	}

	return res, nil
}
