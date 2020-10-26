package listeningSession

import (
	"github.com/47-11/spotifete/database/model"
	"time"
)

func PollSessions() {
	go pollSessions()
}

func pollSessions() {
	for range time.Tick(5 * time.Second) {
		activeSessions := FindFullListeningSessions(model.SimpleListeningSession{
			Active: true,
		})

		for _, session := range activeSessions {
			UpdateSessionIfNecessary(session)
		}
	}
}
