package main

import (
	"database/sql"
	"log"
	"net"

	"github.com/lh-khanhduy/banco_de_rata/api"
	db "github.com/lh-khanhduy/banco_de_rata/db/sqlc"
	"github.com/lh-khanhduy/banco_de_rata/pb"
	"github.com/lh-khanhduy/banco_de_rata/services"
	"github.com/lh-khanhduy/banco_de_rata/utils"
	_ "github.com/lib/pq"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	config, err := utils.LoadConfig(".")
	if err != nil {
		log.Fatal("cannot load configuration: ", err)
	}

	conn, err := sql.Open(config.DBDriver, config.DBSource)
	if err != nil {
		log.Fatal("cannot connect to database: ", err)
	}

	store := db.NewStore(conn)
	runGRPCServer(config, store)
}

// runGinServer will run a HTTP server
func runGinServer(config utils.Config, store db.Store) {
	server, err := api.NewServer(config, store)
	if err != nil {
		log.Fatal("cannot create server: ", err)
	}

	err = server.Start(config.HTTPServerAddress)
	if err != nil {
		log.Fatal("cannot star server: ", err)
	}
}

// runGRPCServer will run a gRPC server
func runGRPCServer(config utils.Config, store db.Store) {
	server, err := services.NewServer(config, store)
	if err != nil {
		log.Fatal("cannot create server: ", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterUserServicesServer(grpcServer, server)
	reflection.Register(grpcServer)

	listener, err := net.Listen("tcp", config.GRPCServerAddress)
	if err != nil {
		log.Fatal("cannot create listener: ", err)
	}

	log.Printf("starting gRPC server at %s", listener.Addr().String())
	err = grpcServer.Serve(listener)
	if err != nil {
		log.Fatal("cannot start gRPC server: ", err)
	}
}
