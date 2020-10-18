package listeningSession

import (
	"github.com/47-11/spotifete/database/model"
	. "github.com/47-11/spotifete/error"
	"github.com/47-11/spotifete/model/dto"
	"github.com/zmb3/spotify"
	"net/http"
	"strings"
)

func SearchTrack(client spotify.Client, query string, limit int) ([]dto.TrackMetadataDto, *SpotifeteError) {
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

func SearchPlaylist(client spotify.Client, query string, limit int) ([]dto.PlaylistMetadataDto, *SpotifeteError) {
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
