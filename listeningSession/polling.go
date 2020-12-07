package listeningSession

import (
	"fmt"
	"github.com/47-11/spotifete/database/model"
	"github.com/google/logger"
	"time"
)

func StartPollSessionsLoop() {
	go pollSessionsLoop()
}

func pollSessionsLoop() {
	for range time.Tick(5 * time.Second) {
		go pollSessions()
	}
}

func pollSessions() {
	logger.Info("Polling sessions...")

	activeSessions := FindFullListeningSessions(model.SimpleListeningSession{
		Active: true,
	})
	logger.Info(fmt.Sprintf("Found %d active sessions.", len(activeSessions)))

	for _, session := range activeSessions {
		UpdateSessionIfNecessary(session)
	}

	logger.Info("Finished polling sessions.")
}
