package db

import (
	log "github.com/sirupsen/logrus"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func NewDb(dsn string) (db *gorm.DB, err error) {
	return gorm.Open(postgres.Open(dsn), &gorm.Config{})
}

func AutoMigrate(db *gorm.DB) error {
	models := []interface{}{
		&Application{},
		&SshTask{},
		&HttpTask{},
	}
	for _, model := range models {
		log.Debugf("Auto migrating model %T", model)
		err := db.AutoMigrate(model)
		if err != nil {
			return err
		}
	}
	return nil
}
