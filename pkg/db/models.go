package db

import (
	"gorm.io/datatypes"
	"gorm.io/gorm"
	"time"
)

const (
	TaskTypeSsh int = iota
	TaskTypeHttp
)

type User struct {
	gorm.Model
	Username    string `gorm:"uniqueIndex"`
	HashedToken string
	LastUsedAt  *time.Time
	Role        string
}

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
	HttpTask      *HttpTask
}

type SshTask struct {
	gorm.Model
	TaskId   uint
	Username string
	Host     string
	Port     uint
	Command  string
}

type HttpTask struct {
	gorm.Model
	TaskId  uint
	Method  string
	Url     string
	Headers datatypes.JSONMap
	Body    string
}
