package db

import (
	"gorm.io/gorm"
	"time"
)

type Application struct {
	gorm.Model
	Name           string
	Description    string
	Secret         string
	LatestVersion  string
	LatestCommit   string
	LastDeployedAt time.Time
}

type HttpTask struct {
	gorm.Model
	Endpoint string
	Method   string
}
