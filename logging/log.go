package logging

import (
	"fmt"
	"github.com/47-11/spotifete/config"
	"github.com/google/logger"
	"os"
	"path/filepath"
	"time"
)

var logDirectory = config.Get().SpotifeteConfiguration.LogDirectory

func SetupLogging() {
	setupSpotifeteLog()
	setupSentryLog()
	setupGinLog()
}

func OpenLogFile(logFileName string) *os.File {
	logFilePath := filepath.Join(logDirectory, logFileName)

	err := os.MkdirAll(filepath.Dir(logFilePath), os.ModePerm)
	if err != nil {
		logger.Fatalf("Could not create log file directory: %v", err)
	}

	moveOldLogFileIfNecessary(logFilePath)

	return doOpenLogFile(logFilePath)
}

func moveOldLogFileIfNecessary(logFilePath string) {
	backupFilePath := fmt.Sprintf("%s-%s.old", logFilePath, time.Now().Round(time.Second).Format(time.RFC3339))

	err := os.Rename(logFilePath, backupFilePath)
	if err != nil && !os.IsNotExist(err) {
		logger.Fatalf("Could not move old error log file: %v", err)
	}
}

func doOpenLogFile(logFilePath string) *os.File {
	openFile, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, os.ModePerm)
	if err != nil {
		logger.Fatalf("Failed to open log file: %v", err)
	}

	return openFile
}
