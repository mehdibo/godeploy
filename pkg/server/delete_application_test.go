package server

import (
	"github.com/mehdibo/godeploy/pkg/db"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
	"net/http"
	"testing"
)

func (s *ServerTestSuite) TestDeleteApplication() {
	s.T().Run("unauthenticated", func(t *testing.T) {
		ctx, rec := prepareRequest(http.MethodDelete, "/api/applications/1", nil, nil)
		if assert.NoError(t, s.server.DeleteApplication(ctx, 1)) {
			assert.Equal(t, http.StatusForbidden, rec.Code)
		}
	})
	s.T().Run("not found", func(t *testing.T) {
		ctx, rec := prepareRequest(http.MethodDelete, "/api/applications/20", nil, &adminUser)
		if assert.NoError(t, s.server.DeleteApplication(ctx, 20)) {
			assert.Equal(t, http.StatusNotFound, rec.Code)
		}
	})
	s.T().Run("existing id", func(t *testing.T) {
		ctx, rec := prepareRequest(http.MethodDelete, "/api/applications/1", nil, &adminUser)
		if assert.NoError(t, s.server.DeleteApplication(ctx, 1)) {
			assert.Equal(t, http.StatusNoContent, rec.Code)
			var app db.Application
			var tasks []db.Task
			tx := s.tx.First(&app, 1)
			assert.Error(t, tx.Error)
			assert.Equal(t, gorm.ErrRecordNotFound, tx.Error)
			_ = s.tx.Find(&tasks, "application_id", 1)
			assert.Empty(t, tasks)
		}
	})
}
