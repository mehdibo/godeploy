package main

import (
	"encoding/json"
	"errors"
	"github.com/mehdibo/go_deploy/pkg/db"
	"github.com/mehdibo/go_deploy/pkg/deployer"
	"github.com/mehdibo/go_deploy/pkg/env"
	"github.com/mehdibo/go_deploy/pkg/messenger"
	log "github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
	"gorm.io/gorm"
	"time"
)

const (
	// MaxAttempts maximum attempts for a failed message
	MaxAttempts = 5
)

func getDb() (*gorm.DB, error) {
	// Load database credentials
	dbHost := env.Get("DB_HOST")
	dbUser := env.Get("DB_USER")
	dbPass := env.Get("DB_PASS")
	dbName := env.Get("DB_NAME")
	dbPort := env.GetDefault("DB_PORT", "5432")
	if dbHost == "" || dbUser == "" || dbPass == "" || dbName == "" {
		return nil, errors.New("required database credentials are not set, check your .env file")
	}
	dsn := "host=" + dbHost + " user=" + dbUser + " password=" + dbPass + " dbname=" + dbName + " port=" + dbPort
	dbConn, err := db.NewDb(dsn)
	if err != nil {
		return nil, err
	}
	log.Info("Running auto migrations")
	err = db.AutoMigrate(dbConn)
	if err != nil {
		return nil, err
	}
	return dbConn, nil
}

func getMessenger() (*messenger.Messenger, error) {
	// Load broker credentials
	brHost := env.Get("AMQP_HOST")
	brUser := env.Get("AMQP_USER")
	brPass := env.Get("AMQP_PASS")
	brPort := env.GetDefault("AMQP_PORT", "5672")
	if brHost == "" || brUser == "" || brPass == "" {
		return nil, errors.New("required broker credentials are not set, check your .env file")
	}
	return messenger.NewMessenger("amqp://" + brUser + ":" + brPass + "@" + brHost + ":" + brPort + "/")
}

func ackMsg(d amqp.Delivery) {
	log.Info("Acknowledging message")
	err := d.Ack(false)
	if err != nil {
		log.Errorf("Couldn't acknowledge message: %s", err.Error())
	}
}

func main() {
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
	})
	log.Info("Loading .env files")
	env.LoadDotEnv()
	logLvl := log.DebugLevel
	if env.Get("APP_ENV") == "prod" {
		logLvl = log.InfoLevel
	}
	log.SetLevel(logLvl)

	log.Info("Connecting to database")
	orm, err := getDb()
	if err != nil {
		log.Fatalf("Couldn't get database : %s", err.Error())
	}

	log.Info("Connecting to AMQP broker")
	msn, err := getMessenger()
	if err != nil {
		log.Fatalf("Couldn't get messenger %s", err.Error())
	}

	dply := deployer.NewDeployer()

	msgs, ch, err := msn.GetMessages(messenger.AppDeployQueue)
	if err != nil {
		log.Fatalf("Failed to get messages: %s", err.Error())
	}
	defer ch.Close()

	var forever chan struct{}

	go func() {
		for d := range msgs {
			log.Info("Received a message")
			log.Debugf("Message body: %s", d.Body)

			var msgBody map[string]uint
			err := json.Unmarshal(d.Body, &msgBody)
			if err != nil {
				log.Errorf("Couldn't decode message body: %s", err.Error())
				continue
			}
			appId, exists := msgBody["id"]
			if !exists {
				ackMsg(d)
				log.Errorf("Invalid payload")
				continue
			}
			appAttempt, exists := msgBody["attempt"]
			if !exists {
				ackMsg(d)
				log.Errorf("Invalid payload")
				continue
			}
			log.Infof("Attempt %d out of %d", appAttempt, MaxAttempts)
			var app db.Application

			tx := orm.Preload("Tasks.HttpTask").Preload("Tasks.SshTask").First(&app, appId)
			if tx.Error != nil {
				if tx.Error == gorm.ErrRecordNotFound {
					ackMsg(d)
					log.Error("Application not found")
					continue
				}
				log.Errorf("Database error: %s", tx.Error.Error())
				continue
			}

			err = dply.DeployApp(&app)
			if err != nil {
				if err == deployer.ErrRecoverable {
					if appAttempt >= MaxAttempts {
						ackMsg(d)
						log.Errorf("Reached maximum attempts")
						continue
					}
					body, err := json.Marshal(map[string]uint{
						"id":      appId,
						"attempt": appAttempt + 1,
					})
					if err != nil {
						log.Errorf("Couldn't marshal payload: %s", err)
						continue
					}
					log.Warning("Failed with a recoverable error, postponing message")
					err = msn.Publish(messenger.AppDeployQueue, body)
					if err != nil {
						log.Errorf("Failed to publish message: %s", err.Error())
					}
					time.Sleep(3 * time.Second)
				}
				if err == deployer.ErrUnrecoverable {
					log.Error("Deployment failed with unrecoverable message")
				}
			}
			ackMsg(d)
		}
	}()

	log.Printf("Waiting for messages. To exit press CTRL+C")
	<-forever
}
