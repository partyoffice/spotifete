package service

import (
	"github.com/47-11/spotifete/config"
	"github.com/47-11/spotifete/database"
	. "github.com/47-11/spotifete/model/database"
	"github.com/47-11/spotifete/model/dto"
	"github.com/jinzhu/gorm"
	"github.com/zmb3/spotify"
	"golang.org/x/oauth2"
	"sync"
)

type spotifyService struct {
	authenticator *spotify.Authenticator
}

var spotifyServiceInstance *spotifyService
var spotifyServiceOnce sync.Once

func SpotifyService() *spotifyService {
	spotifyServiceOnce.Do(func() {
		c := config.GetConfig()
		callbackUrl := c.GetString("server.baseUrl") + "/spotify/callback"

		newAuth := spotify.NewAuthenticator(callbackUrl, spotify.ScopePlaylistReadPrivate, spotify.ScopePlaylistModifyPrivate, spotify.ScopeUserLibraryRead, spotify.ScopeUserModifyPlaybackState, spotify.ScopeUserReadCurrentlyPlaying)
		newAuth.SetAuthInfo(c.GetString("spotify.id"), c.GetString("spotify.secret"))

		spotifyServiceInstance = &spotifyService{
			authenticator: &newAuth,
		}
	})
	return spotifyServiceInstance
}

func (s spotifyService) GetAuthenticator() spotify.Authenticator {
	return *s.authenticator
}

func (s spotifyService) NewAuthUrl() (string, string) {
	sessionId := LoginSessionService().newSessionId()
	database.Connection.Create(&LoginSession{
		Model:     gorm.Model{},
		SessionId: sessionId,
		UserId:    nil,
		Active:    true,
	})
	return s.GetAuthenticator().AuthURL(sessionId), sessionId
}

func (s spotifyService) CheckTokenValidity(token *oauth2.Token) (bool, error) {
	client := s.GetAuthenticator().NewClient(token)
	user, err := client.CurrentUser()
	if err != nil && user == nil {
		// TODO actually verify that the token is invalid and not some other error occurred
		return false, err
	} else {
		return true, nil
	}
}

func (s spotifyService) SearchTrack(client spotify.Client, query string, limit int) ([]dto.TrackMetadataDto, error) {
	result, err := client.SearchOpt(query, spotify.SearchTypeTrack, &spotify.Options{
		Limit: &limit,
	})
	if err != nil {
		return nil, err
	}

	var resultDtos []dto.TrackMetadataDto
	for _, track := range result.Tracks.Tracks {
		metadata, err := s.AddOrUpdateTrackMetadata(client, track.ID)
		if err != nil {
			return nil, err
		}

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
		newTrack := TrackMetadata{
			SpotifyTrackId: trackId.String(),
		}

		newTrack.SetMetadata(*spotifyTrack)

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

func (s spotifyService) GetTrackMetadataById(id uint) *TrackMetadata {
	var foundTracks = []TrackMetadata{}
	database.Connection.Where(TrackMetadata{Model: gorm.Model{
		ID: id,
	}}).Find(&foundTracks)

	if len(foundTracks) > 0 {
		return &foundTracks[0]
	} else {
		return nil
	}
}
