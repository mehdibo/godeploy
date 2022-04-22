package server

import (
	"github.com/labstack/echo/v4"
	"github.com/mehdibo/godeploy/pkg/api"
	"github.com/mehdibo/godeploy/pkg/auth"
	"github.com/mehdibo/godeploy/pkg/db"
	"net/http"
	"sort"
	"strings"
)

func getHttpTasks(ctx echo.Context, rawTasks []api.NewHttpTask) ([]db.Task, error) {
	var tasks []db.Task
	var prevPriority = -100
	for _, httpTask := range rawTasks {
		var task db.Task
		var newHttpTask db.HttpTask

		newHttpTask.Method = strings.ToUpper(httpTask.Method)
		newHttpTask.Url = httpTask.Url

		if httpTask.Headers != nil {
			for _, val := range *httpTask.Headers {
				if _, isString := val.(string); !isString {
					return nil, &echo.HTTPError{
						Code:    http.StatusBadRequest,
						Message: "HTTP header values must all be of the type string",
					}
				}
			}
			newHttpTask.Headers = *(httpTask.Headers)
		}

		if httpTask.Body != nil {
			newHttpTask.Body = *(httpTask.Body)
		}

		task.Priority = uint(httpTask.Priority)
		if httpTask.Priority == prevPriority {
			task.Priority++
		}
		prevPriority = int(task.Priority)
		task.TaskType = db.TaskTypeHttp
		task.HttpTask = &newHttpTask

		if err := ctx.Validate(newHttpTask); err != nil {
			return nil, err
		}
		if err := ctx.Validate(task); err != nil {
			return nil, err
		}

		tasks = append(tasks, task)
	}
	return tasks, nil
}

func getSshTasks(ctx echo.Context, rawTasks []api.NewSshTask) ([]db.Task, error) {
	var tasks []db.Task
	var prevPriority = -100
	for _, sshTask := range rawTasks {
		var task db.Task
		var newSshTask db.SshTask

		newSshTask.ServerFingerprint = sshTask.Fingerprint
		newSshTask.Username = sshTask.Username
		newSshTask.Host = sshTask.Host
		newSshTask.Port = uint(sshTask.Port)
		newSshTask.Command = sshTask.Command

		task.Priority = uint(sshTask.Priority)
		if sshTask.Priority == prevPriority {
			task.Priority++
		}
		prevPriority = sshTask.Priority

		task.TaskType = db.TaskTypeSsh
		task.SshTask = &newSshTask

		if err := ctx.Validate(newSshTask); err != nil {
			return nil, err
		}
		if err := ctx.Validate(task); err != nil {
			return nil, err
		}

		tasks = append(tasks, task)
	}
	return tasks, nil
}

func (srv *Server) AddApplication(ctx echo.Context) error {
	if !isGranted(ctx, auth.RoleAdmin) {
		return accessForbidden(ctx)
	}
	newApp := new(api.NewApplication)
	if err := ctx.Bind(newApp); err != nil {
		return err
	}
	// Convert API structs to DB models
	application := new(db.Application)
	application.Name = newApp.Name
	if newApp.Description != nil {
		application.Description = *(newApp.Description)
	}

	// Generate deployment secret
	rawSecret, err := auth.GenerateToken()
	if err != nil {
		return err
	}
	application.Secret = auth.HashToken(rawSecret)

	// Extract tasks
	var tasks []db.Task

	if newApp.HttpTasks != nil {
		newHttpTasks, err := getHttpTasks(ctx, *(newApp.HttpTasks))
		if err != nil {
			return err
		}
		tasks = append(tasks, newHttpTasks...)
	}

	if newApp.SshTasks != nil {
		newSshTasks, err := getSshTasks(ctx, *(newApp.SshTasks))
		if err != nil {
			return err
		}
		tasks = append(tasks, newSshTasks...)
	}

	sort.SliceStable(tasks, func(i, j int) bool {
		return tasks[i].Priority < tasks[j].Priority
	})
	application.Tasks = tasks

	if err := ctx.Validate(application); err != nil {
		return err
	}

	srv.db.Create(&application)

	// Prepare response
	createdApp := api.CreatedApplication{
		Description: &application.Description,
		Id:          int(application.ID),
		Name:        application.Name,
		RawSecret:   rawSecret,
	}
	return ctx.JSON(http.StatusCreated, createdApp)
}
