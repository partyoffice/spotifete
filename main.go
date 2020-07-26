package main

import (
	"github.com/47-11/spotifete/config"
	"github.com/47-11/spotifete/database"
	"github.com/47-11/spotifete/service"
	"github.com/47-11/spotifete/webapp"
	"github.com/getsentry/sentry-go"
	"github.com/google/logger"
	"log"
	"os"
)

func main() {
	setupLogger()
	setupConfiguration()
	setupSentryIfNeccessary()
	setupDatabase()

	go service.ListeningSessionService().PollSessions()
	webapp.Initialize()
}

func setupLogger() {
	logFile, err := os.OpenFile("spotifete.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0660)
	if err != nil {
		logger.Fatalf("Failed to open log file: %v", err)
	}
	defer logFile.Close()
	defer logger.Init("spotifete", true, false, logFile).Close()
	logger.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
}

func setupConfiguration() {
	c := config.Get()
	if c.SpotifeteConfiguration.ReleaseMode {
		logger.Info("Starting SpotiFete in release mode...")
	} else {
		logger.Warning("Starting SpotiFete in debug mode! To enable release mode, set server.releaseMode to true in config file.")
	}
}

func setupSentryIfNeccessary() {
	if shouldUseSentry() {
		logger.Info("Initializing sentry...")
		setupSentry()
		logger.Info("Sentry initialization successful.")
	} else {
		logger.Warning("Skipping sentry initialization!")
	}
}

func shouldUseSentry() bool {
	configuration := config.Get()
	return configuration.SpotifeteConfiguration.ReleaseMode && configuration.SentryConfiguration.Dsn != nil
}

func setupSentry() {
	configuration := config.Get()
	err := sentry.Init(configuration.SentryConfiguration.GetSentryClientOptions())

	if err != nil {
		logger.Fatalf("Sentry initialization failed: " + err.Error())
	}
}

func setupDatabase() {
	// We want to run migrations etc. before starting the webapp
	database.GetConnection()

	// Close database connection on shutdown
	defer database.CloseConnection()
}
