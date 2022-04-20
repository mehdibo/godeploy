package server

import (
	"encoding/json"
	"github.com/labstack/echo/v4"
	"github.com/mehdibo/go_deploy/pkg/api"
	"github.com/mehdibo/go_deploy/pkg/auth"
	"github.com/mehdibo/go_deploy/pkg/db"
	"github.com/mehdibo/go_deploy/pkg/messenger"
	"net/http"
)

func (srv *Server) DeployApplication(ctx echo.Context, id int) error {
	var app db.Application
	res := srv.db.First(&app, id)
	if res.RowsAffected == 0 {
		return ctx.NoContent(http.StatusNotFound)
	}
	payload := new(api.TriggerDeployment)
	if err := ctx.Bind(payload); err != nil {
		return err
	}
	// Verify secret
	if app.Secret != auth.HashToken(payload.Secret) {
		return accessForbidden(ctx)
	}
	// Check if version is already deployed
	if payload.Version != nil && app.LatestVersion == *payload.Version {
		return badRequest(ctx, "This version is already deployed")
	}
	// Add deployment to queue
	body, err := json.Marshal(map[string]uint{
		"id": app.ID,
	})
	if err != nil {
		return err
	}
	err = srv.msn.Publish(messenger.AppDeployQueue, body)
	if err != nil {
		return err
	}
	return ctx.NoContent(http.StatusOK)
}
