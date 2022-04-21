package env

import (
	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
	"os"
)

// Get fetch the env var or an empty string if not set
func Get(key string) string {
	return os.Getenv(key)
}

// GetDefault fetch the env var or an defaultValue if not set
func GetDefault(key string, defaultValue string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		return defaultValue
	}
	return value
}

// LoadDotEnv load .env files in the following order:
// .env, .env.local, .env.$APP_ENV, .env.$APP_ENV.local
// If there is a problem loading any .env file, it will just be ignored
// TODO: check for required values
func LoadDotEnv() {
	_ = godotenv.Load(".env")
	appEnv := os.Getenv("APP_ENV")
	if appEnv == "" {
		appEnv = "dev"
		_ = os.Setenv("APP_ENV", "dev")
	}
	files := []string{
		".env.local",
		".env." + appEnv + "",
		".env." + appEnv + ".local",
	}
	for _, file := range files {
		log.Debugf("Loading %s file", file)
		_ = godotenv.Overload(file)
	}
}
