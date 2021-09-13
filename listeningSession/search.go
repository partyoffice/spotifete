package listeningSession

import (
	"net/http"
	"strings"

	"github.com/partyoffice/spotifete/database/model"
	. "github.com/partyoffice/spotifete/shared"
	"github.com/partyoffice/spotifete/users"
	"github.com/zmb3/spotify"
)

func SearchTrack(listeningSession model.FullListeningSession, query string, limit int) ([]model.TrackMetadata, *SpotifeteError) {
	client := Client(listeningSession)
	return searchTrack(*client, query, limit)
}

func searchTrack(client spotify.Client, query string, limit int) ([]model.TrackMetadata, *SpotifeteError) {
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

	var resultMetadata []model.TrackMetadata
	for _, track := range result.Tracks.Tracks {
		metadata := model.TrackMetadata{}.SetMetadata(track)
		resultMetadata = append(resultMetadata, metadata)
	}

	return resultMetadata, nil
}

func SearchPlaylist(listeningSession model.FullListeningSession, query string, limit int) ([]model.PlaylistMetadata, *SpotifeteError) {
	client := users.Client(listeningSession.Owner)
	return searchPlaylist(*client, query, limit)
}

func searchPlaylist(client spotify.Client, query string, limit int) ([]model.PlaylistMetadata, *SpotifeteError) {
	cleanedQuery := strings.TrimSpace(query) + "*"
	result, err := client.SearchOpt(cleanedQuery, spotify.SearchTypePlaylist, &spotify.Options{
		Limit: &limit,
	})
	if err != nil {
		return nil, NewError("Could not search for track on Spotify.", err, http.StatusInternalServerError)
	}

	var resultMetadata []model.PlaylistMetadata
	for _, playlist := range result.Playlists.Playlists {
		metadata := model.PlaylistMetadata{}.FromSimplePlaylist(playlist)
		resultMetadata = append(resultMetadata, metadata)
	}

	return resultMetadata, nil
}
