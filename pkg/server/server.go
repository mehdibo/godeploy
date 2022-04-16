package server

import (
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
	"net/http"
)

// Server Represents a server to handle requests, must implement GoDeploy.ServerInterface
type Server struct {
	db *gorm.DB
}

// NewServer create a Server instance
func NewServer(db *gorm.DB) *Server {
	return &Server{db: db}
}

// Ping returns a simple JSON payload to test the server
func (srv Server) Ping(ctx echo.Context) error {
	return ctx.JSON(http.StatusOK, map[string]string{
		"message": "pong",
	})
}
