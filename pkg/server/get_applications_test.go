package server

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

func (s *ServerTestSuite) TestGetApplications() {
	s.T().Run("unauthenticated", func(t *testing.T) {
		ctx, rec := prepareRequest(http.MethodGet, "/api/applications", nil, nil)
		if assert.NoError(t, s.server.AddApplication(ctx)) {
			assert.Equal(t, http.StatusForbidden, rec.Code)
		}
	})
	s.T().Run("valid request", func(t *testing.T) {
		ctx, rec := prepareRequest(http.MethodGet, "/api/application", nil, &adminUser)
		if assert.NoError(t, s.server.GetApplications(ctx)) {
			var resp map[string]interface{}
			err := json.Unmarshal(rec.Body.Bytes(), &resp)
			if assert.NoError(t, err) {
				assert.Len(t, resp["items"], 2)
			}
		}
	})
}
