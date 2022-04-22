package db

import (
	"gorm.io/datatypes"
	"gorm.io/gorm"
	"time"
)

type TaskType int

const (
	TaskTypeSsh TaskType = iota
	TaskTypeHttp
)

func (t TaskType) String() string {
	return [...]string{"SshTask", "HttpTask"}[t]
}

func (t TaskType) EnumIndex() int {
	return int(t)
}

type User struct {
	gorm.Model
	Username    string `gorm:"uniqueIndex"`
	HashedToken string
	LastUsedAt  *time.Time
	Role        string
}

type Application struct {
	gorm.Model
	Name           string `validate:"required"`
	Description    string
	Secret         string
	LatestVersion  string
	LatestCommit   string
	LastDeployedAt time.Time
	Tasks          []Task `validate:"required"`
}

type Task struct {
	gorm.Model
	ApplicationId uint
	Priority      uint
	TaskType      TaskType
	SshTask       *SshTask
	HttpTask      *HttpTask
}

type SshTask struct {
	gorm.Model
	TaskId            uint
	ServerFingerprint string `validate:"required,fingerprint"`
	Username          string `validate:"required"`
	Host              string `validate:"required"`
	Port              uint   `validate:"required"`
	Command           string `validate:"required"`
}

type HttpTask struct {
	gorm.Model
	TaskId  uint
	Method  string `validate:"required"`
	Url     string `validate:"required,url"`
	Headers datatypes.JSONMap
	Body    string
}
