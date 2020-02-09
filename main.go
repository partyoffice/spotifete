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
	// Setup logger
	logFile, err := os.OpenFile("spotifete.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0660)
	if err != nil {
		logger.Fatalf("Failed to open log file: %v", err)
	}
	defer logFile.Close()
	defer logger.Init("spotifete", true, false, logFile).Close()
	logger.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	releaseMode := config.GetConfig().GetBool("spotifete.releaseMode")
	if releaseMode {
		logger.Info("Starting SpotiFete in release mode...")
	} else {
		logger.Warning("Starting SpotiFete in debug mode! To enable release mode, set server.releaseMode to true in config file.")
	}

	// Initialize sentry
	if releaseMode && config.GetConfig().IsSet("sentry.dsn") {
		logger.Info("Initializing sentry...")

		err := sentry.Init(sentry.ClientOptions{
			Dsn:              config.GetConfig().GetString("sentry.dsn"),
			AttachStacktrace: true,
			IgnoreErrors:     []string{".*The access token expired.*", ".*Refresh token revoked.*"},
		})

		if err != nil {
			logger.Fatalf("Sentry initialization failed: " + err.Error())
		} else {
			logger.Info("Sentry initialization successful.")
		}
	} else {
		logger.Warning("Skipping sentry initialization!")
	}

	// Initialize database connection
	database.GetConnection()
	defer database.Shutdown()

	// Start polling sessions
	go service.ListeningSessionService().PollSessions()

	webapp.Initialize()
}
