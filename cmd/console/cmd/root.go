package cmd

import (
	"errors"
	"github.com/mehdibo/godeploy/pkg/db"
	"github.com/mehdibo/godeploy/pkg/env"
	"github.com/spf13/cobra"
	"gorm.io/gorm"
)

var (
	orm     *gorm.DB
	Version = "dev-version"
)

// NewRootCmd create the root command
func NewRootCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "console",
		Short: "Console to manage Go Deploy",
	}
}

// rootCmd represents the root command
var rootCmd = NewRootCmd()

func Execute() error {

	return rootCmd.Execute()
}

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
	err = db.AutoMigrate(dbConn)
	if err != nil {
		return nil, err
	}
	return dbConn, nil
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags()
	rootCmd.Version = Version
}

func initConfig() {
	var err error
	env.LoadDotEnv()
	orm, err = getDb()
	if err != nil {
		panic(err)
	}
}
