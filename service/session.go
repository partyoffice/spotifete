package service

import (
	"github.com/47-11/spotifete/database"
	. "github.com/47-11/spotifete/model"
)

type SessionService struct{}

var db = database.GetInstance()

func (s SessionService) GetActiveSessions() []Session {
	var sessions []Session
	db.Find(&sessions, "active = true")
	return sessions
}

func (s SessionService) GetSessionById(id int64) (Session, error) {
	var sessions []Session
	db.Where("id = ?", id).Limit(1).Find(&sessions)

	if len(sessions) == 1 {
		return sessions[0], nil
	} else {
		return Session{}, EntryNotFoundError{Message: "Session not found."}
	}
}
