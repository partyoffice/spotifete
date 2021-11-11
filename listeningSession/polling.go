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

	sessionsToUpdate := findSessionsToUpdate()

	for _, session := range sessionsToUpdate {
		UpdateSessionIfNecessary(session)
	}
}

func findSessionsToUpdate() []model.FullListeningSession {

	activeSessions := FindFullListeningSessions(model.SimpleListeningSession{
		Active: true,
	})

	recentlyUpdatedThreshold := time.Now().Add(-1 * time.Hour)

	var recentlyUpdatedSessions []model.FullListeningSession
	for _, s := range activeSessions {
		if s.UpdatedAt.After(recentlyUpdatedThreshold) {
			recentlyUpdatedSessions = append(recentlyUpdatedSessions, s)
		}
	}

	return recentlyUpdatedSessions
}
