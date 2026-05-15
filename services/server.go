package services

import (
	"fmt"

	db "github.com/lh-khanhduy/banco_de_rata/db/sqlc"
	"github.com/lh-khanhduy/banco_de_rata/pb"
	"github.com/lh-khanhduy/banco_de_rata/token"
	"github.com/lh-khanhduy/banco_de_rata/utils"
)

// Server serves gRPC request for banking services
type Server struct {
	pb.UnimplementedUserServiceServer
	pb.UnimplementedAccountServiceServer
	config     utils.Config
	store      db.Store
	tokenMaker token.Maker
}

// NewServer creates a gRPC server
func NewServer(config utils.Config, store db.Store) (*Server, error) {
	// use Paseto for generating token
	tokenMaker, err := token.NewPasetoMaker(config.TokenSymmetricKey)
	if err != nil {
		return nil, fmt.Errorf("cannot create tokenMaker: %w", err)
	}

	server := &Server{
		config:     config,
		store:      store,
		tokenMaker: tokenMaker,
	}

	return server, nil
}
