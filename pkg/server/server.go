package server

import (
	"crypto/subtle"
	"github.com/labstack/echo/v4"
	"github.com/mehdibo/go_deploy/pkg/auth"
	"github.com/mehdibo/go_deploy/pkg/db"
	"gorm.io/gorm"
	"net/http"
	"time"
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
func (srv *Server) Ping(ctx echo.Context) error {
	return ctx.JSON(http.StatusOK, map[string]string{
		"message": "pong",
	})
}

// ValidateBasicAuth validate basic auth credentials, used with built-in middleware
func (srv *Server) ValidateBasicAuth(username string, rawToken string, ctx echo.Context) (bool, error) {
	var user db.User
	// Get user
	tx := srv.db.First(&user, "username = ?", username)
	if tx.RowsAffected != 1 {
		return false, nil
	}
	// Verify token
	providedToken := auth.HashToken(rawToken)
	if subtle.ConstantTimeCompare([]byte(providedToken), []byte(user.HashedToken)) == 1 {
		srv.db.Model(&user).Update("LastUsedAt", time.Now())
		ctx.Set(auth.UserKey, user)
		return true, nil
	}
	return false, nil
}
