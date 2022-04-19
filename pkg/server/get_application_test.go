package server

import (
	"encoding/json"
	"github.com/mehdibo/go_deploy/pkg/api"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
	"time"
)

func (s *ServerTestSuite) TestGetApplication() {
	s.T().Run("unauthenticated", func(t *testing.T) {
		ctx, rec := prepareRequest(http.MethodGet, "/api/applications", nil, nil)
		if assert.NoError(t, s.server.GetApplication(ctx, 1)) {
			assert.Equal(t, http.StatusForbidden, rec.Code)
		}
	})
	s.T().Run("not found", func(t *testing.T) {
		ctx, rec := prepareRequest(http.MethodGet, "/api/applications/20", nil, &adminUser)
		if assert.NoError(t, s.server.GetApplication(ctx, 20)) {
			assert.Equal(t, http.StatusNotFound, rec.Code)
		}
	})
	s.T().Run("existing id", func(t *testing.T) {
		ctx, rec := prepareRequest(http.MethodGet, "/api/applications/1", nil, &adminUser)
		if assert.NoError(t, s.server.GetApplication(ctx, 1)) {
			var app api.ApplicationItem
			assert.Equal(t, http.StatusOK, rec.Code)
			err := json.Unmarshal(rec.Body.Bytes(), &app)
			assert.NoError(t, err)
			assert.Equal(t, "Some app to test with", *(app.Description))
			assert.Equal(t, 1, app.Id)
			assert.IsType(t, &time.Time{}, app.LastDeployedAt)
			assert.Equal(t, "", *app.LatestCommit)
			assert.Equal(t, "", *app.LatestVersion)
			assert.Equal(t, "Test App 1", app.Name)

			assert.Len(t, *app.Tasks, 2)

			assert.Equal(t, 0, (*app.Tasks)[0].Priority)
			assert.Equal(t, api.TaskItemTaskTypeHttpTask, (*app.Tasks)[0].TaskType)
			assert.IsType(t, map[string]interface{}{}, (*app.Tasks)[0].Task)
			assert.Equal(t, "", (*app.Tasks)[0].Task.(map[string]interface{})["body"])
			assert.Nil(t, (*app.Tasks)[0].Task.(map[string]interface{})["headers"])
			assert.Equal(t, http.MethodGet, (*app.Tasks)[0].Task.(map[string]interface{})["method"])
			assert.Equal(t, "https://example.com", (*app.Tasks)[0].Task.(map[string]interface{})["url"])

			assert.Equal(t, 1, (*app.Tasks)[1].Priority)
			assert.Equal(t, api.TaskItemTaskTypeSshTask, (*app.Tasks)[1].TaskType)
			assert.IsType(t, map[string]interface{}{}, (*app.Tasks)[1].Task)
			assert.Equal(t, "/update.sh", (*app.Tasks)[1].Task.(map[string]interface{})["command"])
			assert.Equal(t, "localhost", (*app.Tasks)[1].Task.(map[string]interface{})["host"])
			assert.Equal(t, float64(22), (*app.Tasks)[1].Task.(map[string]interface{})["port"])
			assert.Equal(t, "spoody", (*app.Tasks)[1].Task.(map[string]interface{})["username"])
		}
	})
}
