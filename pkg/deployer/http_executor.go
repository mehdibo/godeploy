package deployer

import (
	"github.com/mehdibo/go_deploy/pkg/db"
	log "github.com/sirupsen/logrus"
	"io"
	"net/http"
	"strings"
)

func ExecuteHttpTask(task *db.HttpTask) error {
	var body io.Reader = nil
	if task.Body != "" {
		body = strings.NewReader(task.Body)
	}
	req, err := http.NewRequest(task.Method, task.Url, body)
	if err != nil {
		log.Errorf("Couldn't create request: %s", err.Error())
		return ErrUnrecoverable
	}
	for headerName, headerVal := range task.Headers {
		var val string
		switch v := headerVal.(type) {
		case string:
			val = v
		default:
			log.Errorf("Invalid header %s, all headers must be of the type string", headerName)
			return ErrUnrecoverable
		}
		req.Header.Set(headerName, val)
	}
	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Errorf("An error occurred when sending the request: %s", err.Error())
		return ErrRecoverable
	}
	if resp.StatusCode > http.StatusBadRequest {
		log.Errorf("Server returned status: %d %s", resp.StatusCode, resp.Status)
		return ErrRecoverable
	}
	return nil
}
