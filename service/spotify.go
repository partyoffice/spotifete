package service

import (
	"github.com/47-11/spotifete/authentication"
	"github.com/47-11/spotifete/database"
	"github.com/47-11/spotifete/database/model"
	. "github.com/47-11/spotifete/error"
	"github.com/47-11/spotifete/model/dto"
	"github.com/zmb3/spotify"
	"net/http"
	"strings"
	"sync"
)

type spotifyService struct {
	Clients map[string]*spotify.Client
}

var spotifyServiceInstance *spotifyService
var spotifyServiceOnce sync.Once

func SpotifyService() *spotifyService {
	spotifyServiceOnce.Do(func() {
		spotifyServiceInstance = &spotifyService{
			Clients: map[string]*spotify.Client{},
		}
	})
	return spotifyServiceInstance
}

func (s spotifyService) GetClientForSpotifyUser(spotifyUserId string) *spotify.Client {
	if client, ok := s.Clients[spotifyUserId]; ok {
		s.refreshAndSaveTokenForSpotifyUserIfNeccessary(*client, spotifyUserId)
		return client
	}

	user := UserService().GetUserBySpotifyId(spotifyUserId)
	return s.GetClientForUser(*user)
}

func (s spotifyService) GetClientForUser(user model.User) *spotify.Client {
	if client, ok := s.Clients[user.SpotifyId]; ok {
		s.refreshAndSaveTokenForUserIfNeccessary(*client, user)
		return client
	}

	token := user.GetToken()
	if token == nil {
		return nil
	}

	client := authentication.GetSpotifyAuthenticator().NewClient(token)
	s.refreshAndSaveTokenForUserIfNeccessary(client, user)
	s.Clients[user.SpotifyId] = &client

	return &client
}

func (s spotifyService) refreshAndSaveTokenForSpotifyUserIfNeccessary(client spotify.Client, spotifyUserId string) {
	user := UserService().GetUserBySpotifyId(spotifyUserId)
	s.refreshAndSaveTokenForUserIfNeccessary(client, *user)
}

func (s spotifyService) refreshAndSaveTokenForUserIfNeccessary(client spotify.Client, user model.User) *SpotifeteError {
	newToken, err := client.Token() // This should refresh the token if neccessary: https://github.com/zmb3/spotify/issues/108#issuecomment-568899119
	if err != nil {
		return NewError("Could not refresh Spotify access token. Please try to log out and log in again.", err, http.StatusUnauthorized)
	}

	if newToken.Expiry.After(user.SpotifyTokenExpiry) {
		// Token was updated, persist to database
		// Do this in a goroutine so API calls don't have to wait for the database write to succeed
		go UserService().SetToken(user, *newToken)
	}

	return nil
}

func (s spotifyService) SearchTrack(client spotify.Client, query string, limit int) ([]dto.TrackMetadataDto, *SpotifeteError) {
	cleanedQuery := strings.TrimSpace(query) + "*"

	currentUser, err := client.CurrentUser()
	if err != nil {
		return nil, NewError("Could not fetch information on session owner from Spotify.", err, http.StatusInternalServerError)
	}

	result, err := client.SearchOpt(cleanedQuery, spotify.SearchTypeTrack, &spotify.Options{
		Limit:   &limit,
		Country: &currentUser.Country,
	})
	if err != nil {
		return nil, NewError("Could not search for track on Spotify.", err, http.StatusInternalServerError)
	}

	var resultDtos []dto.TrackMetadataDto
	for _, track := range result.Tracks.Tracks {
		metadata := model.TrackMetadata{}.SetMetadata(track)
		resultDtos = append(resultDtos, dto.TrackMetadataDto{}.FromDatabaseModel(metadata))
	}

	return resultDtos, nil
}

func (s spotifyService) SearchPlaylist(client spotify.Client, query string, limit int) ([]dto.PlaylistMetadataDto, *SpotifeteError) {
	cleanedQuery := strings.TrimSpace(query) + "*"
	result, err := client.SearchOpt(cleanedQuery, spotify.SearchTypePlaylist, &spotify.Options{
		Limit: &limit,
	})
	if err != nil {
		return nil, NewError("Could not search for track on Spotify.", err, http.StatusInternalServerError)
	}

	var resultDtos []dto.PlaylistMetadataDto
	for _, playlist := range result.Playlists.Playlists {
		resultDtos = append(resultDtos, dto.PlaylistMetadataDto{}.FromDatabaseModel(model.PlaylistMetadata{}.FromSimplePlaylist(playlist)))
	}

	return resultDtos, nil
}

func (s spotifyService) AddOrUpdateTrackMetadata(client spotify.Client, spotifyTrack spotify.FullTrack) model.TrackMetadata {
	track := s.GetTrackMetadataBySpotifyTrackId(spotifyTrack.ID.String())
	if track != nil {
		updatedTrack := track.SetMetadata(spotifyTrack)

		database.GetConnection().Save(&updatedTrack)

		return updatedTrack
	} else {
		newTrack := model.TrackMetadata{}.SetMetadata(spotifyTrack)

		database.GetConnection().Create(&newTrack)

		return newTrack
	}
}

func (s spotifyService) GetTrackMetadataBySpotifyTrackId(trackId string) *model.TrackMetadata {
	var foundTracks []model.TrackMetadata
	database.GetConnection().Where(model.TrackMetadata{SpotifyTrackId: trackId}).Find(&foundTracks)

	if len(foundTracks) > 0 {
		return &foundTracks[0]
	} else {
		return nil
	}
}

func (s spotifyService) AddOrUpdatePlaylistMetadata(client spotify.Client, playlistId spotify.ID) (model.PlaylistMetadata, *SpotifeteError) {
	spotifyPlaylist, err := client.GetPlaylist(playlistId)
	if err != nil {
		return model.PlaylistMetadata{}, NewError("Could not get playlist information from Spotify.", err, http.StatusInternalServerError)
	}

	knownPlaylistMetadata := s.GetPlaylistMetadataBySpotifyPlaylistId(playlistId.String())
	if knownPlaylistMetadata != nil {
		updatedPlaylistMetadata := knownPlaylistMetadata.FromFullPlaylist(*spotifyPlaylist)

		database.GetConnection().Save(&updatedPlaylistMetadata)

		return updatedPlaylistMetadata, nil
	} else {
		newPlaylistMetadata := model.PlaylistMetadata{}.FromFullPlaylist(*spotifyPlaylist)

		database.GetConnection().Create(&newPlaylistMetadata)

		return newPlaylistMetadata, nil
	}
}

func (s spotifyService) GetPlaylistMetadataBySpotifyPlaylistId(playlistId string) *model.PlaylistMetadata {
	var foundPlaylists []model.PlaylistMetadata
	database.GetConnection().Where(model.PlaylistMetadata{SpotifyPlaylistId: playlistId}).Find(&foundPlaylists)

	if len(foundPlaylists) > 0 {
		return &foundPlaylists[0]
	} else {
		return nil
	}
}
