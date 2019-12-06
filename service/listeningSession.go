package service

import (
	"errors"
	"fmt"
	"github.com/47-11/spotifete/database"
	database2 "github.com/47-11/spotifete/model/database"
	"github.com/jinzhu/gorm"
	"github.com/zmb3/spotify"
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
	database.Connection.Model(&database2.ListeningSession{}).Count(&count)
	return count
}

func (listeningSessionService) GetActiveSessionCount() int {
	var count int
	database.Connection.Model(&database2.ListeningSession{}).Where(database2.ListeningSession{Active: true}).Count(&count)
	return count
}

func (listeningSessionService) GetActiveSessions() []database2.ListeningSession {
	var sessions []database2.ListeningSession
	database.Connection.Where(database2.ListeningSession{Active: true}).Find(&sessions)
	return sessions
}

func (listeningSessionService) GetSessionById(id uint) *database2.ListeningSession {
	var sessions []database2.ListeningSession
	database.Connection.Where(database2.ListeningSession{Model: gorm.Model{ID: id}}).Find(&sessions)

	if len(sessions) == 1 {
		return &sessions[0]
	} else {
		return nil
	}
}

func (listeningSessionService) GetSessionByJoinId(joinId string) *database2.ListeningSession {
	if len(joinId) == 0 {
		return nil
	}

	var sessions []database2.ListeningSession
	database.Connection.Where(database2.ListeningSession{JoinId: &joinId}).Find(&sessions)

	if len(sessions) == 1 {
		return &sessions[0]
	} else {
		return nil
	}
}

func (listeningSessionService) GetActiveSessionsByOwnerId(ownerId uint) []database2.ListeningSession {
	var sessions []database2.ListeningSession
	database.Connection.Where(database2.ListeningSession{Active: true, OwnerId: ownerId}).Find(&sessions)
	return sessions
}

func (s listeningSessionService) NewSession(user *database2.User, title string) (*database2.ListeningSession, error) {
	client := SpotifyService().GetAuthenticator().NewClient(user.GetToken())

	joinId := s.newJoinId()
	playlist, err := client.CreatePlaylistForUser(user.SpotifyId, fmt.Sprintf("%s - SpotiFete", title), fmt.Sprintf("Automatic playlist for SpotiFete session %s. You can join using the code %s.", title, joinId), false)
	if err != nil {
		return nil, err
	}

	listeningSession := database2.ListeningSession{
		Model:           gorm.Model{},
		Active:          true,
		OwnerId:         user.ID,
		JoinId:          &joinId,
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
	database.Connection.Model(&database2.ListeningSession{}).Where(database2.ListeningSession{JoinId: &joinId}).Count(&count)
	return count > 0
}

func (s listeningSessionService) CloseSession(user *database2.User, joinId string) error {
	session := s.GetSessionByJoinId(joinId)
	if user.ID != session.OwnerId {
		return errors.New("only the owner can close a session")
	}

	session.Active = false
	session.JoinId = nil
	database.Connection.Save(&session)

	client := SpotifyService().authenticator.NewClient(user.GetToken())
	return client.ReplacePlaylistTracks(spotify.ID(session.SpotifyPlaylist))
}
