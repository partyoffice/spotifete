package service

import (
	"errors"
	"fmt"
	"github.com/47-11/spotifete/database"
	. "github.com/47-11/spotifete/model/database"
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
	database.Connection.Model(&ListeningSession{}).Count(&count)
	return count
}

func (listeningSessionService) GetActiveSessionCount() int {
	var count int
	database.Connection.Model(&ListeningSession{}).Where(ListeningSession{Active: true}).Count(&count)
	return count
}

func (listeningSessionService) GetActiveSessions() []ListeningSession {
	var sessions []ListeningSession
	database.Connection.Where(ListeningSession{Active: true}).Find(&sessions)
	return sessions
}

func (listeningSessionService) GetSessionById(id uint) *ListeningSession {
	var sessions []ListeningSession
	database.Connection.Where(ListeningSession{Model: gorm.Model{ID: id}}).Find(&sessions)

	if len(sessions) == 1 {
		return &sessions[0]
	} else {
		return nil
	}
}

func (listeningSessionService) GetSessionByJoinId(joinId string) *ListeningSession {
	if len(joinId) == 0 {
		return nil
	}

	var sessions []ListeningSession
	database.Connection.Where(ListeningSession{JoinId: &joinId}).Find(&sessions)

	if len(sessions) == 1 {
		return &sessions[0]
	} else {
		return nil
	}
}

func (listeningSessionService) GetActiveSessionsByOwnerId(ownerId uint) []ListeningSession {
	var sessions []ListeningSession
	database.Connection.Where(ListeningSession{Active: true, OwnerId: ownerId}).Find(&sessions)
	return sessions
}

func (s listeningSessionService) NewSession(user *User, title string) (*ListeningSession, error) {
	client := SpotifyService().GetAuthenticator().NewClient(user.GetToken())

	joinId := s.newJoinId()
	playlist, err := client.CreatePlaylistForUser(user.SpotifyId, fmt.Sprintf("%s - SpotiFete", title), fmt.Sprintf("Automatic playlist for SpotiFete session %s. You can join using the code %s.", title, joinId), false)
	if err != nil {
		return nil, err
	}

	listeningSession := ListeningSession{
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
	database.Connection.Model(&ListeningSession{}).Where(ListeningSession{JoinId: &joinId}).Count(&count)
	return count > 0
}

func (s listeningSessionService) CloseSession(user *User, joinId string) error {
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

func (s listeningSessionService) RequestSong(session *ListeningSession, trackId string) error {
	sessionOwner := UserService().GetUserById(session.OwnerId)
	client := SpotifyService().GetAuthenticator().NewClient(sessionOwner.GetToken())

	track, err := client.GetTrack(spotify.ID(trackId))
	if err != nil {
		return err
	}

	newSongRequest := SongRequest{
		Model:     gorm.Model{},
		SessionId: session.ID,
		UserId:    nil,
		TrackId:   track.ID.String(),
	}
	database.Connection.Create(&newSongRequest)

	// TODO: For now just add the request to the playlist
	_, err = client.AddTracksToPlaylist(spotify.ID(session.SpotifyPlaylist), track.ID)
	if err != nil {
		return err
	}

	return nil
}
