package service

import (
	"github.com/47-11/spotifete/database"
	. "github.com/47-11/spotifete/model"
)

type ListeningSessionService struct{}

func (s ListeningSessionService) GetTotalSessionCount() int {
	var count int
	database.Connection.Model(&ListeningSession{}).Count(&count)
	return count
}

func (s ListeningSessionService) GetActiveSessionCount() int {
	var count int
	database.Connection.Model(&ListeningSession{}).Where("active = true").Count(&count)
	return count
}

func (s ListeningSessionService) GetActiveSessions() []ListeningSession {
	var sessions []ListeningSession
	database.Connection.Where("active = true").Find(&sessions)
	return sessions
}

func (s ListeningSessionService) GetSessionById(id int64) (ListeningSession, error) {
	var sessions []ListeningSession
	database.Connection.Where("id = ?", id).Find(&sessions)

	if len(sessions) == 1 {
		return sessions[0], nil
	} else {
		return ListeningSession{}, EntryNotFoundError{Message: "Session not found."}
	}
}

func (s ListeningSessionService) GetSessionByJoinId(id uint) (ListeningSession, error) {
	var sessions []ListeningSession
	database.Connection.Where("join_id = ?", id).Find(&sessions)

	if len(sessions) == 1 {
		return sessions[0], nil
	} else {
		return ListeningSession{}, EntryNotFoundError{Message: "Session not found."}
	}
}

func (s ListeningSessionService) GetActiveSessionsByOwnerId(ownerId uint) []ListeningSession {
	var sessions []ListeningSession
	database.Connection.Where("active = true AND owner_id = ?", ownerId).Find(&sessions)
	return sessions
}
