package main

import (
	"infra/api/internal/app"
	"infra/api/internal/config"
	"infra/api/internal/infra/nats"
	"infra/api/internal/infra/postgres"
	"infra/api/internal/logger"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load(os.Getenv("ENVPATH"))
	if err != nil {
		panic("Can't load .env file: " + err.Error())
	}

	config := config.ReadConfig()
	config.DB = postgres.Init(config)

	unixLogger := logger.Init(config)

	natsinfra := nats.Init(config, unixLogger)

	app := &app.App{
		Config:    config,
		Db:        config.DB,
		NatsInfra: natsinfra,
		Log:       unixLogger,
	}

	// go Autostart(appConfig)
	app.Start()

}
