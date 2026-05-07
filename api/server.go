package api

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	db "github.com/lh-khanhduy/banco_de_rata/db/sqlc"
	"github.com/lh-khanhduy/banco_de_rata/token"
	"github.com/lh-khanhduy/banco_de_rata/utils"
)

// Server serves HTTP request for banking services
type Server struct {
	config     utils.Config
	store      db.Store
	tokenMaker token.Maker
	router     *gin.Engine
}

// NewServer creates a HTTP server and setup routing
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

	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterValidation("currency", validCurrency)
	}

	server.setupRouter()

	return server, nil
}

func (s *Server) setupRouter() {
	router := gin.Default()

	// USER
	router.POST("/users", s.createUser)
	router.POST("/users/login", s.loginUser)

	// ACCOUNTS
	router.POST("/accounts", s.createAccount)
	router.GET("/accounts/:id", s.getAccount)
	router.GET("/accounts", s.listAccounts)
	router.PATCH("/accounts", s.updateAccount)
	router.DELETE("/accounts/:id", s.deleteAccount)

	// TRANSFER
	router.POST("/transfers", s.createTransfer)

	s.router = router
}

// Start runs the HTTP server on a specific address
func (s *Server) Start(address string) error {
	return s.router.Run(address)
}

func errorResponse(err error) gin.H {
	return gin.H{"error": err.Error()}
}
