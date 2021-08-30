package listeningSession

import (
	"time"

	"github.com/partyoffice/spotifete/database/model"
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
