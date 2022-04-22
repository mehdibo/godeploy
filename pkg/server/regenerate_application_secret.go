package server

import (
	"github.com/labstack/echo/v4"
	"github.com/mehdibo/godeploy/pkg/api"
	"github.com/mehdibo/godeploy/pkg/auth"
	"github.com/mehdibo/godeploy/pkg/db"
	"gorm.io/gorm"
	"net/http"
)

func (srv *Server) RegenerateApplicationSecret(ctx echo.Context, id int) error {
	if !isGranted(ctx, auth.RoleAdmin) {
		return accessForbidden(ctx)
	}
	var app db.Application
	res := srv.db.First(&app, id)
	if res.Error != nil {
		if res.Error == gorm.ErrRecordNotFound {
			return ctx.NoContent(http.StatusNotFound)
		}
		return errorMsg(ctx, http.StatusInternalServerError, "Something went wrong")
	}
	// Generate new secret
	rawSecret, err := auth.GenerateToken()
	if err != nil {
		return err
	}
	// Save it to database
	app.Secret = auth.HashToken(rawSecret)
	tx := srv.db.Save(&app)
	if tx.Error != nil {
		return errorMsg(ctx, http.StatusInternalServerError, "Something went wrong")
	}
	// Output it
	return ctx.JSON(http.StatusOK, api.CreatedApplication{
		Description: &app.Description,
		Id:          int(app.ID),
		Name:        app.Name,
		RawSecret:   rawSecret,
	})
}
