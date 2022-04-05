package db

import (
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func NewDb(dsn string) (db *gorm.DB, err error) {
	return gorm.Open(postgres.Open(dsn), &gorm.Config{})
}
