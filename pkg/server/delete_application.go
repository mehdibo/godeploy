package server

import (
	"github.com/labstack/echo/v4"
	"github.com/mehdibo/godeploy/pkg/auth"
	"github.com/mehdibo/godeploy/pkg/db"
	"net/http"
)

func (srv *Server) DeleteApplication(ctx echo.Context, id int) error {
	if !isGranted(ctx, auth.RoleAdmin) {
		return accessForbidden(ctx)
	}
	var app db.Application
	res := srv.db.Preload("Tasks.HttpTask").Preload("Tasks.SshTask").First(&app, id)
	if res.RowsAffected == 0 {
		return ctx.NoContent(http.StatusNotFound)
	}
	for _, task := range app.Tasks {
		if task.TaskType == db.TaskTypeSsh {
			tx := srv.db.Delete(task.SshTask)
			if tx.Error != nil {
				return tx.Error
			}
		}
		if task.TaskType == db.TaskTypeHttp {
			tx := srv.db.Delete(task.HttpTask)
			if tx.Error != nil {
				return tx.Error
			}
		}
		tx := srv.db.Delete(&task)
		if tx.Error != nil {
			return tx.Error
		}
	}
	tx := srv.db.Delete(&app)
	if tx.Error != nil {
		return tx.Error
	}
	return ctx.NoContent(http.StatusNoContent)
}
