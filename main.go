package main

import (
	"github.com/47-11/spotifete/config"
	"github.com/47-11/spotifete/database"
	"github.com/47-11/spotifete/listeningSession"
	"github.com/47-11/spotifete/webapp"
	"github.com/getsentry/sentry-go"
	"github.com/google/logger"
	"log"
	"os"
)

var logFile *os.File
var spotifeteWebapp webapp.SpotifeteWebapp

func main() {
	defer shutdown()
	setup()
	run()
}

func setup() {
	setupLogger()
	config.Get()
	setupSentryIfNeccessary()
	database.GetConnection()
	setupWebapp()
}

func setupLogger() {
	logFile = openLogFile()
	logger.Init("spotifete", true, false, logFile)
	logger.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
}

func openLogFile() *os.File {
	openFile, err := os.OpenFile("spotifete.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0660)
	if err != nil {
		logger.Fatalf("Failed to open log file: %v", err)
	}

	return openFile
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
		logger.Fatalf("Sentry initialization failed: %s" + err.Error())
	}
}

func setupWebapp() {
	spotifeteWebapp = webapp.SpotifeteWebapp{}.Setup()
}

func run() {
	go listeningSession.PollSessions()
	spotifeteWebapp.Run()
}

func shutdown() {
	defer closeLogFile()
	spotifeteWebapp.Shutdown()
}

func closeLogFile() {
	err := logFile.Close()
	if err != nil {
		panic("Could not close log file: " + err.Error())
	}
}
