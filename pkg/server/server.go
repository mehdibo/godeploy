package server

import (
	"crypto/subtle"
	"github.com/labstack/echo/v4"
	"github.com/mehdibo/go_deploy/pkg/auth"
	"github.com/mehdibo/go_deploy/pkg/db"
	log "github.com/sirupsen/logrus"
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

func isGranted(ctx echo.Context, role string) bool {
	user, err := auth.LoadUserFromCtx(ctx)
	if err != nil {
		if err == auth.ErrUserTypeMismatch {
			log.Errorf("Failed to load user: %s", err.Error())
		}
		return false
	}
	if user.Role == role {
		return true
	}
	return false
}

func accessForbidden(ctx echo.Context) error {
	return ctx.JSON(http.StatusForbidden, map[string]string{
		"message": "Access forbidden.",
	})
}

// Ping returns a simple JSON payload to test the server
func (srv *Server) Ping(ctx echo.Context) error {
	if !isGranted(ctx, auth.RoleAdmin) {
		return accessForbidden(ctx)
	}
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
