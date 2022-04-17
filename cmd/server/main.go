package main

import (
	"errors"
	"github.com/labstack/echo/v4"
	mdl "github.com/labstack/echo/v4/middleware"
	"github.com/mehdibo/go_deploy/pkg/api"
	"github.com/mehdibo/go_deploy/pkg/db"
	"github.com/mehdibo/go_deploy/pkg/env"
	"github.com/mehdibo/go_deploy/pkg/middleware"
	"github.com/mehdibo/go_deploy/pkg/server"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"net/http"
	"time"
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

	e := echo.New()

	e.Static("/assets", "swagger-ui/assets")
	e.File("/docs", "swagger-ui/index.html")

	e.Use(middleware.RequestLog)
	e.Use(mdl.TimeoutWithConfig(mdl.TimeoutConfig{
		Timeout: 30 * time.Second,
	}))

	g := e.Group("/api")
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

	srv := server.NewServer(orm)
	api.RegisterHandlers(g, srv)

	e.Logger.Fatal(e.Start(":8080"))
}
