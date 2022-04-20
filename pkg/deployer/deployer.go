package deployer

import (
	"errors"
	"github.com/mehdibo/go_deploy/pkg/db"
	log "github.com/sirupsen/logrus"
)

var (
	ErrRecoverable   = errors.New("an error occurred but a retry might solve it")
	ErrUnrecoverable = errors.New("an error occurred and a retry will not solve the problem")
)

type Deployer struct {
}

func NewDeployer() *Deployer {
	return &Deployer{}
}

func (d *Deployer) DeployApp(app *db.Application) error {
	// Loop through tasks
	for _, task := range app.Tasks {
		// Pass each task to the appropriate task executor
		switch task.TaskType {
		case db.TaskTypeSsh:
			log.Info("Executing SSH task")
		case db.TaskTypeHttp:
			log.Info("Executing HTTP task")
			err := ExecuteHttpTask(task.HttpTask)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
