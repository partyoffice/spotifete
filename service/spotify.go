package service

import (
	"github.com/47-11/spotifete/config"
	"github.com/47-11/spotifete/database"
	. "github.com/47-11/spotifete/model/database"
	"github.com/47-11/spotifete/model/dto"
	"github.com/jinzhu/gorm"
	"github.com/zmb3/spotify"
	"golang.org/x/oauth2"
	"strings"
	"sync"
)

type spotifyService struct {
	Authenticator spotify.Authenticator
	Clients       map[string]*spotify.Client
}

var spotifyServiceInstance *spotifyService
var spotifyServiceOnce sync.Once

func SpotifyService() *spotifyService {
	spotifyServiceOnce.Do(func() {
		c := config.GetConfig()
		callbackUrl := c.GetString("spotifete.baseUrl") + "/spotify/callback"

		newAuth := spotify.NewAuthenticator(callbackUrl, spotify.ScopePlaylistReadPrivate, spotify.ScopePlaylistModifyPrivate, spotify.ScopeUserLibraryRead, spotify.ScopeUserModifyPlaybackState, spotify.ScopeUserReadCurrentlyPlaying)
		newAuth.SetAuthInfo(c.GetString("spotify.id"), c.GetString("spotify.secret"))

		spotifyServiceInstance = &spotifyService{
			Authenticator: newAuth,
			Clients:       map[string]*spotify.Client{},
		}
	})
	return spotifyServiceInstance
}

func (s spotifyService) GetClientForSpotifyUser(spotifyUserId string) *spotify.Client {
	if client, ok := s.Clients[spotifyUserId]; ok {
		return client
	}

	user := UserService().GetUserBySpotifyId(spotifyUserId)
	return s.GetClientForUser(*user)
}

func (s spotifyService) GetClientForUser(user User) *spotify.Client {
	if client, ok := s.Clients[user.SpotifyId]; ok {
		return client
	}

	token := user.GetToken()
	if token == nil {
		return nil
	}

	client := s.Authenticator.NewClient(token)
	s.Clients[user.SpotifyId] = &client

	return &client
}

func (s spotifyService) NewAuthUrl() (string, string) {
	sessionId := LoginSessionService().newSessionId()
	database.Connection.Create(&LoginSession{
		Model:     gorm.Model{},
		SessionId: sessionId,
		UserId:    nil,
		Active:    true,
	})
	return s.Authenticator.AuthURL(sessionId), sessionId
}

func (s spotifyService) CheckTokenValidity(token *oauth2.Token) (bool, error) {
	client := s.Authenticator.NewClient(token)
	user, err := client.CurrentUser()
	if err != nil && user == nil {
		// TODO actually verify that the token is invalid and not some other error occurred
		return false, err
	} else {
		return true, nil
	}
}

func (s spotifyService) SearchTrack(client spotify.Client, query string, limit int) ([]dto.TrackMetadataDto, error) {
	cleanedQuery := strings.TrimSpace(query) + "*"
	result, err := client.SearchOpt(cleanedQuery, spotify.SearchTypeTrack, &spotify.Options{
		Limit: &limit,
	})
	if err != nil {
		return nil, err
	}

	var resultDtos []dto.TrackMetadataDto
	for _, track := range result.Tracks.Tracks {
		metadata := TrackMetadata{}.SetMetadata(track)
		resultDtos = append(resultDtos, dto.TrackMetadataDto{}.FromDatabaseModel(metadata))
	}

	return resultDtos, nil
}

func (s spotifyService) AddOrUpdateTrackMetadata(client spotify.Client, trackId spotify.ID) (TrackMetadata, error) {
	spotifyTrack, err := client.GetTrack(trackId)
	if err != nil {
		return TrackMetadata{}, err
	}

	track := s.GetTrackMetadataBySpotifyTrackId(trackId.String())
	if track != nil {
		updatedTrack := track.SetMetadata(*spotifyTrack)

		database.Connection.Save(&updatedTrack)

		return updatedTrack, nil
	} else {
		newTrack := TrackMetadata{}.SetMetadata(*spotifyTrack)

		database.Connection.Create(&newTrack)

		return newTrack, nil
	}
}

func (s spotifyService) GetTrackMetadataBySpotifyTrackId(trackId string) *TrackMetadata {
	var foundTracks = []TrackMetadata{}
	database.Connection.Where(TrackMetadata{SpotifyTrackId: trackId}).Find(&foundTracks)

	if len(foundTracks) > 0 {
		return &foundTracks[0]
	} else {
		return nil
	}
}
