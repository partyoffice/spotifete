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

	c := config.Get()
	if c.SpotifeteConfiguration.ReleaseMode {
		logger.Info("Starting SpotiFete in release mode...")
	} else {
		logger.Warning("Starting SpotiFete in debug mode! To enable release mode, set server.releaseMode to true in config file.")
	}

	// Initialize sentry
	if c.SpotifeteConfiguration.ReleaseMode && c.SentryConfiguration.Dsn != nil {
		logger.Info("Initializing sentry...")

		err := sentry.Init(sentry.ClientOptions{
			Dsn:              *c.SentryConfiguration.Dsn,
			AttachStacktrace: true,
			IgnoreErrors:     []string{".*Refresh token revoked.*"}, // TODO:
		})

		if err != nil {
			logger.Fatalf("Sentry initialization failed: " + err.Error())
		} else {
			logger.Info("Sentry initialization successful.")
		}
	} else {
		logger.Warning("Skipping sentry initialization!")
	}

	// Close database connection on shutdown
	defer database.CloseConnection()

	// Start polling sessions
	go service.ListeningSessionService().PollSessions()

	webapp.Initialize()
}
