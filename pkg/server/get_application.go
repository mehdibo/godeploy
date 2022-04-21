package server

import (
	"github.com/labstack/echo/v4"
	"github.com/mehdibo/godeploy/pkg/api"
	"github.com/mehdibo/godeploy/pkg/auth"
	"github.com/mehdibo/godeploy/pkg/db"
	"net/http"
)

func (srv *Server) GetApplication(ctx echo.Context, id int) error {
	if !isGranted(ctx, auth.RoleAdmin) {
		return accessForbidden(ctx)
	}
	var app db.Application
	res := srv.db.Preload("Tasks.HttpTask").Preload("Tasks.SshTask").First(&app, id)
	if res.RowsAffected == 0 {
		return ctx.NoContent(http.StatusNotFound)
	}
	var appItem api.ApplicationItem

	appItem.Description = &app.Description
	appItem.Id = int(app.ID)
	appItem.LastDeployedAt = &app.LastDeployedAt
	appItem.LatestCommit = &app.LatestCommit
	appItem.LatestVersion = &app.LatestVersion
	appItem.Name = app.Name

	var tasks []api.TaskItem
	for _, task := range app.Tasks {
		var taskItem api.TaskItem
		taskItem.Priority = int(task.Priority)
		if task.TaskType == db.TaskTypeSsh {
			var sshTask api.SshTaskItem

			sshTask.Host = task.SshTask.Host
			sshTask.Username = task.SshTask.Username
			sshTask.Command = task.SshTask.Command
			sshTask.Port = int(task.SshTask.Port)

			taskItem.TaskType = api.TaskItemTaskTypeSshTask
			taskItem.Task = sshTask
		}
		if task.TaskType == db.TaskTypeHttp {
			var httpTask api.HttpTaskItem

			httpTask.Url = task.HttpTask.Url
			httpTask.Method = task.HttpTask.Method
			httpTask.Body = &task.HttpTask.Body
			httpTask.Headers = (*map[string]interface{})(&task.HttpTask.Headers)

			taskItem.TaskType = api.TaskItemTaskTypeHttpTask
			taskItem.Task = httpTask
		}
		tasks = append(tasks, taskItem)
	}
	appItem.Tasks = &tasks
	return ctx.JSON(http.StatusOK, appItem)
}
