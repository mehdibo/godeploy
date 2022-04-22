package server

import (
	"github.com/mehdibo/godeploy/pkg/db"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

func (s *ServerTestSuite) TestRegenerateApplicationSecret() {
	s.T().Run("unauthenticated", func(t *testing.T) {
		ctx, rec := prepareRequest(http.MethodGet, "/api/applications/1/regenerate", nil, nil)
		if assert.NoError(t, s.server.RegenerateApplicationSecret(ctx, 1)) {
			assert.Equal(t, http.StatusForbidden, rec.Code)
		}
	})
	s.T().Run("non existing application", func(t *testing.T) {
		ctx, rec := prepareRequest(http.MethodGet, "/api/applications/100/regenerate", nil, &adminUser)
		if assert.NoError(t, s.server.RegenerateApplicationSecret(ctx, 100)) {
			assert.Equal(t, http.StatusNotFound, rec.Code)
		}
	})
	s.T().Run("existing application", func(t *testing.T) {
		id := 1
		var oldApp db.Application
		s.tx.First(&oldApp, id)
		ctx, rec := prepareRequest(http.MethodGet, "/api/applications/1/regenerate", nil, &adminUser)
		if assert.NoError(t, s.server.RegenerateApplicationSecret(ctx, id)) {
			var newApp db.Application
			s.tx.First(&newApp, id)
			assert.Equal(t, http.StatusOK, rec.Code)
			assert.NotEqual(t, oldApp.Secret, newApp.Secret)
		}
	})
}
