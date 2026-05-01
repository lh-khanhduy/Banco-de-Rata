package main

import (
	"database/sql"
	"log"

	"github.com/lh-khanhduy/banco_de_rata/api"
	db "github.com/lh-khanhduy/banco_de_rata/db/sqlc"
	_ "github.com/lib/pq"
)

const (
	dbDriver = "postgres"
	dbSource = "postgresql://root:khanhduy@localhost:5432/banco_de_rata?sslmode=disable"
	address  = "0.0.0.0:8080"
)

func main() {
	conn, err := sql.Open(dbDriver, dbSource)
	if err != nil {
		log.Fatal("cannot connect to database: ", err)
	}

	store := db.NewStore(conn)
	server := api.NewServer(store)

	err = server.Start(address)
	if err != nil {
		log.Fatal("cannot star server: ", err)
	}
}
