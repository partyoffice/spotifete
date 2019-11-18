package service

import (
	"github.com/47-11/spotifete/database"
	"github.com/47-11/spotifete/database/model"
	. "github.com/47-11/spotifete/model"
)

type ListeningSessionService struct{}

func (s ListeningSessionService) GetTotalSessionCount() int {
	var count int
	database.Connection.Model(&model.ListeningSession{}).Count(&count)
	return count
}

func (s ListeningSessionService) GetActiveSessionCount() int {
	var count int
	database.Connection.Model(&model.ListeningSession{}).Where("active = true").Count(&count)
	return count
}

func (s ListeningSessionService) GetActiveSessions() []model.ListeningSession {
	var sessions []model.ListeningSession
	database.Connection.Where("active = true").Find(&sessions)
	return sessions
}

func (s ListeningSessionService) GetSessionById(id int64) (model.ListeningSession, error) {
	var sessions []model.ListeningSession
	database.Connection.Where("id = ?", id).Find(&sessions)

	if len(sessions) == 1 {
		return sessions[0], nil
	} else {
		return model.ListeningSession{}, EntryNotFoundError{Message: "Session not found."}
	}
}

func (s ListeningSessionService) GetSessionByJoinId(id uint) (model.ListeningSession, error) {
	var sessions []model.ListeningSession
	database.Connection.Where("join_id = ?", id).Find(&sessions)

	if len(sessions) == 1 {
		return sessions[0], nil
	} else {
		return model.ListeningSession{}, EntryNotFoundError{Message: "Session not found."}
	}
}

func (s ListeningSessionService) GetActiveSessionsByOwnerId(ownerId uint) []model.ListeningSession {
	var sessions []model.ListeningSession
	database.Connection.Where("active = true AND owner_id = ?", ownerId).Find(&sessions)
	return sessions
}
