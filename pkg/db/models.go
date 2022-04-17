package db

import (
	"gorm.io/gorm"
	"time"
)

const (
	TaskTypeSsh int = iota
	TaskTypeHttp
)

type Application struct {
	gorm.Model
	Name           string
	Description    string
	Secret         string
	LatestVersion  string
	LatestCommit   string
	LastDeployedAt time.Time
	Tasks          []Task
}

type Task struct {
	gorm.Model
	ApplicationId uint
	Priority      uint
	TaskType      int
	SshTask       *SshTask
}

type SshTask struct {
	gorm.Model
	TaskId   uint
	Username string
	Host     string
	Port     uint
	Command  string
}
