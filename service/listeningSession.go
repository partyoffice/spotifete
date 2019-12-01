package service

import (
	"github.com/47-11/spotifete/database"
	"github.com/47-11/spotifete/database/model"
	"github.com/jinzhu/gorm"
	"sync"
)

type listeningSessionService struct{}

var listeningSessionServiceInstance *listeningSessionService
var listeningSessionServiceOnce sync.Once

func ListeningSessionService() *listeningSessionService {
	listeningSessionServiceOnce.Do(func() {
		listeningSessionServiceInstance = &listeningSessionService{}
	})
	return listeningSessionServiceInstance
}

func (listeningSessionService) GetTotalSessionCount() int {
	var count int
	database.Connection.Model(&model.ListeningSession{}).Count(&count)
	return count
}

func (listeningSessionService) GetActiveSessionCount() int {
	var count int
	database.Connection.Model(&model.ListeningSession{}).Where(model.ListeningSession{Active: true}).Count(&count)
	return count
}

func (listeningSessionService) GetActiveSessions() []model.ListeningSession {
	var sessions []model.ListeningSession
	database.Connection.Where(model.ListeningSession{Active: true}).Find(&sessions)
	return sessions
}

func (listeningSessionService) GetSessionById(id uint) *model.ListeningSession {
	var sessions []model.ListeningSession
	database.Connection.Where(model.ListeningSession{Model: gorm.Model{ID: id}}).Find(&sessions)

	if len(sessions) == 1 {
		return &sessions[0]
	} else {
		return nil
	}
}

func (listeningSessionService) GetSessionByJoinId(joinId string) *model.ListeningSession {
	var sessions []model.ListeningSession
	database.Connection.Where(model.ListeningSession{JoinId: joinId}).Find(&sessions)

	if len(sessions) == 1 {
		return &sessions[0]
	} else {
		return nil
	}
}

func (listeningSessionService) GetActiveSessionsByOwnerId(ownerId uint) []model.ListeningSession {
	var sessions []model.ListeningSession
	database.Connection.Where(model.ListeningSession{Active: true, OwnerId: ownerId}).Find(&sessions)
	return sessions
}
