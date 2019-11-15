package service

import (
	"github.com/47-11/spotifete/database"
	. "github.com/47-11/spotifete/model"
)

type SessionService struct{}

func (s SessionService) GetTotalSessionCount() int {
	var count int
	database.Connection.Model(&Session{}).Count(&count)
	return count
}

func (s SessionService) GetActiveSessionCount() int {
	var count int
	database.Connection.Model(&Session{}).Where("active = true").Count(&count)
	return count
}

func (s SessionService) GetActiveSessions() []Session {
	var sessions []Session
	database.Connection.Where("active = true").Find(&sessions)
	return sessions
}

func (s SessionService) GetSessionById(id int64) (Session, error) {
	var sessions []Session
	database.Connection.Where("id = ?", id).Limit(1).Find(&sessions)

	if len(sessions) == 1 {
		return sessions[0], nil
	} else {
		return Session{}, EntryNotFoundError{Message: "Session not found."}
	}
}
