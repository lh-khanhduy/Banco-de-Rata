package services

import (
	"context"
	"database/sql"
	"fmt"

	db "github.com/lh-khanhduy/banco_de_rata/db/sqlc"
	"github.com/lh-khanhduy/banco_de_rata/pb"
	"github.com/lib/pq"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
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

func (s *Server) GetAccount(ctx context.Context, req *pb.GetAccountRequest) (*pb.GetAccountResponse, error) {
	account, err := s.store.GetAccount(ctx, req.GetId())
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Error(codes.NotFound, "account not found")
		}

		return nil, status.Error(codes.Internal, "cannot find account")
	}

	md := s.extractMetadata(ctx)
	if md.Username == "" {
		return nil, status.Errorf(codes.Unauthenticated, "missing or invalid username")
	}

	if account.Owner != md.Username {
		return nil, status.Error(codes.Unauthenticated, "account doesn't belong to the authenticated user")
	}

	res := &pb.GetAccountResponse{
		Account: &pb.Account{
			Id:        account.ID,
			Owner:     account.Owner,
			Currency:  account.Currency,
			Balance:   account.Balance,
			CreatedAt: timestamppb.New(account.CreatedAt),
		},
	}

	return res, nil
}

func (s *Server) ListAccounts(ctx context.Context, req *pb.ListAccountsRequest) (*pb.ListAccountResponse, error) {
	md := s.extractMetadata(ctx)
	args := db.ListAccountParams{
		Owner:  md.Username,
		Limit:  req.PageSize,
		Offset: (req.PageId - 1) * req.PageSize,
	}

	listAccounts, err := s.store.ListAccount(ctx, args)
	if err != nil {
		return nil, status.Error(codes.Internal, "cannot get list account")
	}

	res := &pb.ListAccountResponse{
		Accounts: fromDBAccountToPBAccount(listAccounts),
	}

	return res, nil
}

func fromDBAccountToPBAccount(list []db.Account) []*pb.Account {
	var result []*pb.Account

	for _, acc := range list {
		pbAcc := &pb.Account{
			Id:        acc.ID,
			Owner:     acc.Owner,
			Balance:   acc.Balance,
			Currency:  acc.Currency,
			CreatedAt: timestamppb.New(acc.CreatedAt),
		}

		result = append(result, pbAcc)
	}

	return result
}
