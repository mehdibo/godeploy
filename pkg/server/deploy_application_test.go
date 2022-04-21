package server

import (
	"bytes"
	"encoding/json"
	"github.com/mehdibo/godeploy/pkg/messenger"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

func (s *ServerTestSuite) TestDeployApplication() {
	validPayload := map[string]string{
		"commit":  "fd5e2e86",
		"version": "v1.0.0",
		"secret":  "some_secret",
	}
	uri := "/api/applications/1/deploy"
	s.T().Run("invalid token", func(t *testing.T) {
		b, err := json.Marshal(validPayload)
		r := bytes.NewReader(b)
		if assert.NoError(t, err) {
			ctx, rec := prepareRequest(http.MethodPost, uri, r, nil)
			if assert.NoError(t, s.server.DeployApplication(ctx, 1)) {
				assert.Equal(t, http.StatusForbidden, rec.Code)
			}
		}
	})
	s.T().Run("non existing application", func(t *testing.T) {
		validPayload["secret"] = "deploy_token"
		b, err := json.Marshal(validPayload)
		r := bytes.NewReader(b)
		if assert.NoError(t, err) {
			ctx, rec := prepareRequest(http.MethodPost, "/api/applications/200/deploy", r, nil)
			if assert.NoError(t, s.server.DeployApplication(ctx, 200)) {
				assert.Equal(t, http.StatusNotFound, rec.Code)
			}
		}
	})
	s.T().Run("valid token", func(t *testing.T) {
		b, err := json.Marshal(validPayload)
		r := bytes.NewReader(b)
		if assert.NoError(t, err) {
			ctx, rec := prepareRequest(http.MethodPost, uri, r, nil)
			if assert.NoError(t, s.server.DeployApplication(ctx, 1)) {
				assert.Equal(t, http.StatusOK, rec.Code)
				count, err := s.msn.CountMessages(messenger.AppDeployQueue)
				if assert.NoError(t, err) {
					assert.Equal(t, 1, count)
				}
			}
		}
	})
}
