package auth

import (
	"github.com/labstack/echo/v4"
	"github.com/mehdibo/go_deploy/pkg/db"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestGenerateToken(t *testing.T) {
	token, err := GenerateToken()

	assert.Nil(t, err)
	// We multiply by 2 because the result is in hex
	assert.Len(t, token, TokenSize*2)
}

func TestHashToken(t *testing.T) {
	rawToken := "this_is_a_test"
	hashedToken := HashToken(rawToken)

	assert.Equal(t, "c7313ab341785793fe19a8a6caa87043ff1651853360be9762acfa6ae53f9ce0", hashedToken)
}

func TestLoadUserFromCtx(t *testing.T) {
	t.Run("user not found", func(t *testing.T) {
		e := echo.New()
		ctx := e.AcquireContext()

		_, err := LoadUserFromCtx(ctx)
		assert.Equal(t, err, ErrUserNotFound)
	})
	t.Run("type error", func(t *testing.T) {
		e := echo.New()
		ctx := e.AcquireContext()

		ctx.Set(UserKey, "test")

		_, err := LoadUserFromCtx(ctx)
		assert.Equal(t, err, ErrUserTypeMismatch)
	})
	t.Run("user found", func(t *testing.T) {
		now := time.Now()
		user := db.User{
			Username:    "test",
			HashedToken: "somehashedtoken",
			LastUsedAt:  &now,
			Role:        "ROLE_ADMIN",
		}
		e := echo.New()
		ctx := e.AcquireContext()

		ctx.Set(UserKey, user)

		loadedUser, err := LoadUserFromCtx(ctx)
		assert.Nil(t, err)
		assert.Equal(t, user, loadedUser)
	})
}
