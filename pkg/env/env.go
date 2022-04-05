package env

import (
	"github.com/joho/godotenv"
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
func LoadDotEnv() {
	err := godotenv.Load(".env")
	if err != nil {
		panic(err)
	}
	appEnv := os.Getenv("APP_ENV")
	if appEnv == "" {
		_ = os.Setenv("APP_ENV", "dev")
	}
	files := []string{
		".env.local",
		".env." + appEnv + "",
		".env." + appEnv + ".local",
	}
	for _, file := range files {
		_ = godotenv.Overload(file)
	}
}
