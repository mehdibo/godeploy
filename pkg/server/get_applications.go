package server

import (
	"github.com/labstack/echo/v4"
	"github.com/mehdibo/godeploy/pkg/api"
	"github.com/mehdibo/godeploy/pkg/auth"
	"github.com/mehdibo/godeploy/pkg/db"
	"net/http"
)

func (srv *Server) GetApplications(ctx echo.Context) error {
	if !isGranted(ctx, auth.RoleAdmin) {
		return accessForbidden(ctx)
	}
	var dbApps []db.Application
	var apps []api.ApplicationCollectionItem

	srv.db.Find(&dbApps)

	for _, dbApp := range dbApps {
		app := api.ApplicationCollectionItem{
			Description: &dbApp.Description,
			Id:          int(dbApp.ID),
			Name:        dbApp.Name,
		}
		apps = append(apps, app)
	}
	return ctx.JSON(http.StatusOK, api.ApplicationCollection{Items: apps})
}
