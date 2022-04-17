package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"github.com/labstack/echo/v4"
	"github.com/mehdibo/go_deploy/pkg/db"
)

const (
	// UserKey the key used in the context for the user
	UserKey = "Auth.User"
	// RoleAdmin role given to admin users
	RoleAdmin = "Auth.RoleAdmin"
	// TokenSize token size in bytes
	TokenSize = 21
)

var (
	// ErrUserNotFound user not found in context
	ErrUserNotFound = errors.New("user not found in context")

	// ErrUserTypeMismatch failed to retrieve user, type mismatch
	ErrUserTypeMismatch = errors.New("failed to retrieve user, type mismatch")
)

func GenerateToken() (string, error) {
	b := make([]byte, TokenSize)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func HashToken(rawToken string) string {
	h := sha256.New()
	h.Write([]byte(rawToken))

	return hex.EncodeToString(h.Sum(nil))
}

func LoadUserFromCtx(ctx echo.Context) (db.User, error) {
	var (
		user db.User
		ok   bool
	)

	uval := ctx.Get(UserKey)
	if uval == nil {
		return user, ErrUserNotFound
	}

	user, ok = uval.(db.User)
	if !ok {
		return user, ErrUserTypeMismatch
	}

	return user, nil
}
