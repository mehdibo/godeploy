package main

import (
	"github.com/labstack/echo/v4"
	"github.com/mehdibo/go_deploy/pkg/api"
	"github.com/mehdibo/go_deploy/pkg/db"
	"github.com/mehdibo/go_deploy/pkg/env"
	"github.com/mehdibo/go_deploy/pkg/server"
	"gorm.io/gorm"
)

func getDb() (*gorm.DB, error) {
	// Load database credentials
	dbHost := env.Get("DB_HOST")
	dbUser := env.Get("DB_USER")
	dbPass := env.Get("DB_PASS")
	dbName := env.Get("DB_NAME")
	dbPort := env.GetDefault("DB_PORT", "5432")
	if dbHost == "" || dbUser == "" || dbPass == "" || dbName == "" {
		panic("Required database credentials are not set, check your .env file")
	}
	dsn := "host=" + dbHost + " user=" + dbUser + " password=" + dbPass + " dbname=" + dbName + " port=" + dbPort
	dbConn, err := db.NewDb(dsn)
	if err != nil {
		return nil, err
	}
	err = db.AutoMigrate(dbConn)
	if err != nil {
		return nil, err
	}
	return dbConn, nil
}

func main() {
	env.LoadDotEnv()
	orm, err := getDb()
	if err != nil {
		panic("Couldn't get database :" + err.Error())
	}

	e := echo.New()

	srv := server.NewServer(orm)

	g := e.Group("/api")

	api.RegisterHandlers(g, srv)

	e.Logger.Fatal(e.Start(":8080"))
}