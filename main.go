package main

import (
	"github.com/47-11/spotifete/config"
	"github.com/47-11/spotifete/database"
	"github.com/47-11/spotifete/service"
	"github.com/47-11/spotifete/webapp"
	"github.com/getsentry/sentry-go"
	"log"
)

func main() {
	defer database.Shutdown()

	releaseMode := config.GetConfig().GetBool("spotifete.releaseMode")
	if releaseMode {
		log.Println("Starting SpotiFete in release mode...")
	} else {
		log.Println("Starting SpotiFete in debug mode! To enable release mode, set server.releaseMode to true in config file.")
	}

	if releaseMode && config.GetConfig().IsSet("sentry.dsn") {
		log.Println("Initializing sentry...")

		err := sentry.Init(sentry.ClientOptions{
			Dsn: config.GetConfig().GetString("sentry.dsn"),
		})

		if err != nil {
			panic("Sentry initialization failed: " + err.Error())
		} else {
			log.Println("Sentry initialization successful.")
		}
	} else {
		log.Println("Skipping sentry initialization!")
	}

	go service.ListeningSessionService().PollSessions()

	webapp.Initialize()
}
