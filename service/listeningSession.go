package service

import (
	"fmt"
	"github.com/47-11/spotifete/database"
	"github.com/47-11/spotifete/database/model"
	"github.com/jinzhu/gorm"
	"math/rand"
	"sync"
)

type listeningSessionService struct {
	numberRunes []rune
}

var listeningSessionServiceInstance *listeningSessionService
var listeningSessionServiceOnce sync.Once

func ListeningSessionService() *listeningSessionService {
	listeningSessionServiceOnce.Do(func() {
		listeningSessionServiceInstance = &listeningSessionService{
			numberRunes: []rune("0123456789"),
		}
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
	if len(joinId) == 0 {
		return nil
	}

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

func (s listeningSessionService) NewSession(user *model.User, title string) (*model.ListeningSession, error) {
	client := SpotifyService().GetAuthenticator().NewClient(user.GetToken())

	joinId := s.newJoinId()
	playlist, err := client.CreatePlaylistForUser(user.SpotifyId, fmt.Sprintf("%s - SpotiFete", title), fmt.Sprintf("Automatic playlist for SpotiFete session %s. You can join using the code %s.", title, joinId), false)
	if err != nil {
		return nil, err
	}

	listeningSession := model.ListeningSession{
		Model:           gorm.Model{},
		Active:          true,
		OwnerId:         user.ID,
		JoinId:          joinId,
		SpotifyPlaylist: playlist.ID.String(),
		Title:           title,
	}

	database.Connection.Create(&listeningSession)

	return &listeningSession, nil
}

func (s listeningSessionService) newJoinId() string {
	for {
		b := make([]rune, 8)
		for i := range b {
			b[i] = s.numberRunes[rand.Intn(len(s.numberRunes))]
		}
		newJoinId := string(b)

		if !s.joinIdExists(newJoinId) {
			return newJoinId
		}
	}
}

func (listeningSessionService) joinIdExists(joinId string) bool {
	var count uint
	database.Connection.Model(&model.ListeningSession{}).Where(model.ListeningSession{JoinId: joinId}).Count(&count)
	return count > 0
}
