package main

import (
	"errors"
	"github.com/labstack/echo/v4"
	mdl "github.com/labstack/echo/v4/middleware"
	"github.com/mehdibo/godeploy/pkg/api"
	"github.com/mehdibo/godeploy/pkg/db"
	"github.com/mehdibo/godeploy/pkg/env"
	"github.com/mehdibo/godeploy/pkg/messenger"
	"github.com/mehdibo/godeploy/pkg/middleware"
	"github.com/mehdibo/godeploy/pkg/server"
	"github.com/mehdibo/godeploy/pkg/validator"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"net/http"
	"time"
)

var Version = "dev-version"

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

func main() {
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
	})
	log.Infof("Go Deploy - %s", Version)
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
	defer msn.Close()
	srv := server.NewServer(orm, msn)

	e := echo.New()

	e.Validator = validator.NewValidator()

	e.Static("/assets", "swagger-ui/assets")
	e.File("/docs", "swagger-ui/index.html")

	e.Use(middleware.RequestLog)
	e.Use(mdl.TimeoutWithConfig(mdl.TimeoutConfig{
		Timeout: 30 * time.Second,
	}))

	g := e.Group("/api")

	g.Use(mdl.BasicAuthWithConfig(mdl.BasicAuthConfig{
		Skipper: func(c echo.Context) bool {
			return c.Path() == "/api/swagger.json"
		},
		Validator: srv.ValidateBasicAuth,
		Realm:     "",
	}))

	g.GET("/swagger.json", func(ctx echo.Context) error {
		errMsg := map[string]string{
			"message": "Something went wrong",
		}
		swg, err := api.GetSwagger()
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, errMsg)
		}
		json, err := swg.MarshalJSON()
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, errMsg)
		}
		return ctx.JSONBlob(http.StatusOK, json)
	})

	api.RegisterHandlers(g, srv)

	e.Logger.Fatal(e.Start(":" + env.GetDefault("LISTEN_PORT", "8080")))
}
