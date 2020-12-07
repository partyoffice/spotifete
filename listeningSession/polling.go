package listeningSession

import (
	"github.com/47-11/spotifete/database/model"
	"github.com/google/logger"
	"time"
)

func PollSessions() {
	go pollSessions()
}

func pollSessions() {
	for range time.Tick(5 * time.Second) {
		logger.Info("Polling sessions")

		activeSessions := FindFullListeningSessions(model.SimpleListeningSession{
			Active: true,
		})

		for _, session := range activeSessions {
			UpdateSessionIfNecessary(session)
		}
	}
}
