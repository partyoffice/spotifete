package listeningSession

import (
	"github.com/47-11/spotifete/database/model"
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
	activeSessions := FindFullListeningSessions(model.SimpleListeningSession{
		Active: true,
	})

	for _, session := range activeSessions {
		UpdateSessionIfNecessary(session)
	}
}
