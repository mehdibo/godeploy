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
	"os"
	"time"
)

const (
	// MaxAttempts maximum attempts for a failed message
	MaxAttempts = 5
	// SleepTime seconds to sleep after we requeue a failed job
	SleepTime = 3
)

var (
	orm  *gorm.DB
	dply *deployer.Deployer
	msn  *messenger.Messenger
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

func getDeployer() (*deployer.Deployer, error) {
	sshPrivKey := env.Get("SSH_PRIVATE_KEY")
	sshKnownHosts := "./KnownHosts"
	sshPassPhrase := env.GetDefault("SSH_PASSPHRASE", "")
	if sshPrivKey == "" {
		return nil, errors.New("required SSH config is not set")
	}
	f, err := os.OpenFile(sshKnownHosts, os.O_RDWR|os.O_CREATE, 0600)
	if err != nil {
		return nil, err
	}
	_ = f.Close()
	return deployer.NewDeployer(sshPrivKey, sshPassPhrase, sshKnownHosts), nil
}

func consume(d *amqp.Delivery) {
	// Parse msg body
	var msg messenger.DeployApplication
	err := json.Unmarshal(d.Body, &msg)
	if err != nil {
		log.Errorf("Couldn't decode message body: %s", err.Error())
		return
	}
	// Load Application from DB
	log.Info("Loading application from database")
	var app db.Application
	tx := orm.Preload("Tasks.HttpTask").Preload("Tasks.SshTask").First(&app, msg.ID)
	if tx.Error != nil {
		if tx.Error == gorm.ErrRecordNotFound {
			log.Error("Application not found")
			return
		}
		log.Errorf("Database error: %s", tx.Error.Error())
		return
	}

	log.Info("Running deployment tasks")
	log.Infof("Attempt %d out of %d", msg.Attempt, MaxAttempts)

	err = dply.DeployApp(&app)
	if err == nil {
		log.Info("Deployment was successful")
		if msg.Commit != nil {
			app.LatestCommit = *msg.Commit
		}
		if msg.Version != nil {
			app.LatestVersion = *msg.Version
		}
		app.LastDeployedAt = time.Now()
		tx := orm.Save(&app)
		if tx.Error != nil {
			log.Errorf("Failed to update Application: %s", tx.Error.Error())
		}
		return
	}

	if msg.Attempt >= MaxAttempts {
		log.Warning("Reached maximum attempts, cancelling job")
		return
	}

	if err == deployer.ErrUnrecoverable {
		log.Error("Deployment failed with unrecoverable message, cancelling job")
		return
	}

	if err == deployer.ErrUnrecoverable {
		log.Info("Deployment is recoverable, postponing job")
		msg.Attempt++
		body, err := json.Marshal(msg)
		if err != nil {
			log.Errorf("Couldn't marshal payload: %s", err)
			return
		}
		err = msn.Publish(messenger.AppDeployQueue, body)
		if err != nil {
			log.Errorf("Failed to publish message: %s", err.Error())
			return
		}
		log.Debugf("Sleeping for %d to retry job", SleepTime)
		time.Sleep(SleepTime * time.Second)
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

	var err error

	dply, err = getDeployer()
	if err != nil {
		log.Fatalf("Couldn't get deployer : %s", err.Error())
	}

	log.Info("Connecting to database")
	orm, err = getDb()
	if err != nil {
		log.Fatalf("Couldn't get database : %s", err.Error())
	}

	log.Info("Connecting to AMQP broker")
	msn, err = getMessenger()
	if err != nil {
		log.Fatalf("Couldn't get messenger %s", err.Error())
	}

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

			consume(&d)
			log.Info("Acknowledging message")
			err := d.Ack(false)
			if err != nil {
				log.Errorf("Couldn't acknowledge message: %s", err.Error())
			}
		}
	}()

	log.Printf("Waiting for messages. To exit press CTRL+C")
	<-forever
}
