package logging

import (
	"sync"

	"github.com/getsentry/sentry-go"
	"github.com/google/logger"
	"github.com/partyoffice/spotifete/config"
)

var setupSentryLogOnce sync.Once

func setupSentryLog() {
	setupSentryLogOnce.Do(setupSentryLogIfNecessary)
}

func setupSentryLogIfNecessary() {
	if shouldUseSentry() {
		logger.Info("Initializing sentry...")
		doSetupSentryLog()
		logger.Info("Sentry initialization successful.")
	} else {
		logger.Warning("Skipping sentry initialization!")
	}
}

func shouldUseSentry() bool {
	configuration := config.Get()
	return configuration.SpotifeteConfiguration.ReleaseMode && configuration.SentryConfiguration.Dsn != nil
}

func doSetupSentryLog() {
	configuration := config.Get()
	err := sentry.Init(configuration.SentryConfiguration.GetSentryClientOptions())

	if err != nil {
		logger.Fatalf("Sentry initialization failed: %s" + err.Error())
	}
}
