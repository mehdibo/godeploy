package server

import (
	"encoding/json"
	"github.com/labstack/echo/v4"
	"github.com/mehdibo/godeploy/pkg/auth"
	"github.com/mehdibo/godeploy/pkg/db"
	"github.com/stretchr/testify/assert"
	"gorm.io/datatypes"
	"net/http"
	"strings"
	"testing"
)

func traversePath(data map[string]interface{}, fieldPath string, val interface{}) map[string]interface{} {
	path := strings.Split(fieldPath, ".")
	if len(path) > 1 {
		return traversePath(data[path[0]].(map[string]interface{}), strings.Join(path[1:], "."), val)
	}
	data[path[0]] = val
	return data
}

func getInvalidPayload(fieldPath string, val interface{}) string {
	var validPayload = map[string]interface{}{
		"name":        "Test app",
		"description": "Test description",
		"httpTasks": []map[string]interface{}{
			{
				"method":   "POST",
				"priority": 0,
				"url":      "https://google.com",
			},
		},
		"sshTasks": []map[string]interface{}{
			{
				"command":     "rm -rf *",
				"fingerprint": "SHA256:CPOS6R0RfuXkKZqrMm/HyCHBDqXs7mXxsM9MABd17G8",
				"host":        "localhost",
				"port":        1337,
				"priority":    1,
				"username":    "mehdibo",
			},
		},
	}
	validPayload = traversePath(validPayload, fieldPath, val)
	payload, _ := json.Marshal(validPayload)
	return string(payload)
}

func (s *ServerTestSuite) TestAddApplication() {
	s.T().Run("unauthenticated", func(t *testing.T) {
		ctx, rec := prepareRequest(http.MethodPost, "/api/applications", nil, nil)
		if assert.NoError(t, s.server.AddApplication(ctx)) {
			assert.Equal(t, http.StatusForbidden, rec.Code)
		}
	})
	s.T().Run("bad request", func(t *testing.T) {
		var invalidRequests = []string{
			// no tasks
			`{"name": "Some name"}`,
			// Empty name
			getInvalidPayload("name", ""),
			// Empty method
			getInvalidPayload("httpTasks", []map[string]interface{}{
				{
					"method":   "",
					"priority": 0,
					"url":      "https://google.com",
				},
			}),
			// Empty url
			getInvalidPayload("httpTasks", []map[string]interface{}{
				{
					"method":   "POST",
					"priority": 0,
					"url":      "",
				},
			}),
			// Invalid url
			getInvalidPayload("httpTasks", []map[string]interface{}{
				{
					"method":   "POST",
					"priority": 0,
					"url":      "google.com",
				},
			}),
			// Non string header value
			getInvalidPayload("httpTasks", []map[string]interface{}{
				{
					"method":   "POST",
					"priority": 0,
					"url":      "https://google.com",
					"headers": map[string]interface{}{
						"some-header": 21,
					},
				},
			}),
			// Empty ssh fingerprint
			getInvalidPayload("sshTasks", []map[string]interface{}{
				{
					"priority":    0,
					"fingerprint": "",
					"username":    "user",
					"host":        "host",
					"port":        22,
					"command":     "ls",
				},
			}),
			// Invalid ssh fingerprint format
			getInvalidPayload("sshTasks", []map[string]interface{}{
				{
					"priority":    0,
					"fingerprint": "blabla",
					"username":    "user",
					"host":        "host",
					"port":        22,
					"command":     "ls",
				},
			}),
		}
		for _, payload := range invalidRequests {
			r := strings.NewReader(payload)
			ctx, _ := prepareRequest(http.MethodPost, "/api/applications", r, &adminUser)
			err := s.server.AddApplication(ctx)
			if assert.Error(t, err) {
				validationErr := err.(*echo.HTTPError)
				assert.Equal(t, http.StatusBadRequest, validationErr.Code)
				t.Log(validationErr.Message)
			}
		}
	})

	s.T().Run("valid payload", func(t *testing.T) {
		payload := `
	{
	 "name": "Test app",
	 "description": "Test description",
	 "httpTasks": [
	   {
	     "body": "val=a&bod=b",
	     "headers": {
	       "Test header": "value"
		  },
	     "method": "POST",
	     "priority": 0,
	     "url": "https://google.com"
	   },
		{
	     "method": "POST",
	     "priority": 2,
	     "url": "https://google.com"
	   }
	 ],
	 "sshTasks": [
	   {
	     "command": "ls",
		 "fingerprint": "SHA256:1",
	     "host": "localhost",
	     "port": 22,
	     "priority": 1,
	     "username": "spoody"
	   },
		{
	     "command": "rm -rf *",
		 "fingerprint": "SHA256:2",
	     "host": "somehow",
	     "port": 22,
	     "priority": 3,
	     "username": "mehdibo"
	   }
	 ]
	}
	`
		r := strings.NewReader(payload)
		ctx, rec := prepareRequest(http.MethodPost, "/api/application", r, &adminUser)
		if assert.NoError(t, s.server.AddApplication(ctx)) {
			var resp map[string]interface{}
			assert.Equal(t, http.StatusCreated, rec.Code)
			err := json.Unmarshal(rec.Body.Bytes(), &resp)
			assert.NoError(t, err)
			assert.Equal(t, resp["name"], "Test app")
			assert.Equal(t, resp["description"], "Test description")
			assert.Contains(t, resp, "id")
			assert.Contains(t, resp, "rawSecret")

			// Test that the data saved in the db is correct
			var app db.Application
			s.tx.Preload("Tasks.HttpTask").Preload("Tasks.SshTask").First(&app, resp["id"])

			assert.Equal(t, auth.HashToken(resp["rawSecret"].(string)), app.Secret)

			// Test that tasks are in the correct order
			if assert.Len(t, app.Tasks, 4) {
				assert.Equal(t, uint(0), app.Tasks[0].Priority)
				assert.Equal(t, db.TaskTypeHttp, app.Tasks[0].TaskType)
				assert.Nil(t, app.Tasks[0].SshTask)
				assert.NotNil(t, app.Tasks[0].HttpTask)
				assert.Equal(t, http.MethodPost, app.Tasks[0].HttpTask.Method)
				assert.Equal(t, "https://google.com", app.Tasks[0].HttpTask.Url)
				assert.Equal(t, datatypes.JSONMap{"Test header": "value"}, app.Tasks[0].HttpTask.Headers)
				assert.Equal(t, "val=a&bod=b", app.Tasks[0].HttpTask.Body)

				assert.Equal(t, uint(1), app.Tasks[1].Priority)
				assert.Equal(t, db.TaskTypeSsh, app.Tasks[1].TaskType)
				assert.Nil(t, app.Tasks[1].HttpTask)
				assert.NotNil(t, app.Tasks[1].SshTask)
				assert.Equal(t, "SHA256:1", app.Tasks[1].SshTask.ServerFingerprint)
				assert.Equal(t, "spoody", app.Tasks[1].SshTask.Username)
				assert.Equal(t, "localhost", app.Tasks[1].SshTask.Host)
				assert.Equal(t, uint(22), app.Tasks[1].SshTask.Port)
				assert.Equal(t, "ls", app.Tasks[1].SshTask.Command)

				assert.Equal(t, uint(2), app.Tasks[2].Priority)
				assert.Equal(t, db.TaskTypeHttp, app.Tasks[2].TaskType)
				assert.Nil(t, app.Tasks[2].SshTask)
				assert.NotNil(t, app.Tasks[2].HttpTask)
				assert.Equal(t, http.MethodPost, app.Tasks[2].HttpTask.Method)
				assert.Equal(t, "https://google.com", app.Tasks[2].HttpTask.Url)
				assert.Equal(t, datatypes.JSONMap(nil), app.Tasks[2].HttpTask.Headers)
				assert.Equal(t, "", app.Tasks[2].HttpTask.Body)

				assert.Equal(t, uint(3), app.Tasks[3].Priority)
				assert.Equal(t, db.TaskTypeSsh, app.Tasks[3].TaskType)
				assert.Nil(t, app.Tasks[3].HttpTask)
				assert.NotNil(t, app.Tasks[3].SshTask)
				assert.Equal(t, "SHA256:2", app.Tasks[3].SshTask.ServerFingerprint)
				assert.Equal(t, "mehdibo", app.Tasks[3].SshTask.Username)
				assert.Equal(t, "somehow", app.Tasks[3].SshTask.Host)
				assert.Equal(t, uint(22), app.Tasks[3].SshTask.Port)
				assert.Equal(t, "rm -rf *", app.Tasks[3].SshTask.Command)
			}

		}
	})
}
