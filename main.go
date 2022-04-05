package main

import (
	"github.com/joho/godotenv"
	"os"
)

var (
	appEnv string
)

func loadDotEnvFiles() {
	/*
		We load the .env files in the following order:
		.env
		.env.local
		.env.$APP_ENV
		.env.$APP_ENV.local
	*/
	err := godotenv.Load(".env")
	if err != nil {
		panic(err)
	}
	appEnv = os.Getenv("APP_ENV")
	if appEnv == "" {
		appEnv = "dev"
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

func main() {
	loadDotEnvFiles()
}
