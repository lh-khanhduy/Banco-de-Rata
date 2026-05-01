package api

import (
	"github.com/gin-gonic/gin"
	db "github.com/lh-khanhduy/banco_de_rata/db/sqlc"
)

// Server serves HTTP request for banking services
type Server struct {
	store  *db.Store
	router *gin.Engine
}

// NewServer creates a HTTP server and setup routing
func NewServer(store *db.Store) *Server {
	server := &Server{
		store: store,
	}

	router := gin.Default()

	// set up router here
	router.POST("/accounts", server.createAccount)
	router.GET("/accounts/:id", server.getAccount)
	router.GET("/accounts", server.listAccounts)

	server.router = router
	return server
}

// Start runs the HTTP server on a specific address
func (s *Server) Start(address string) error {
	return s.router.Run(address)
}

func errorResponse(err error) gin.H {
	return gin.H{"error": err.Error()}
}
