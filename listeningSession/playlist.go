package listeningSession

import (
	"github.com/47-11/spotifete/database"
	"github.com/47-11/spotifete/database/model"
	. "github.com/47-11/spotifete/shared"
	"github.com/zmb3/spotify"
	"net/http"
)

func AddOrUpdatePlaylistMetadata(client spotify.Client, playlistId spotify.ID) (model.PlaylistMetadata, *SpotifeteError) {
	spotifyPlaylist, err := client.GetPlaylist(playlistId)
	if err != nil {
		return model.PlaylistMetadata{}, NewError("Could not get playlist information from Spotify.", err, http.StatusInternalServerError)
	}

	knownPlaylistMetadata := GetPlaylistMetadataBySpotifyPlaylistId(playlistId.String())
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

func GetPlaylistMetadataBySpotifyPlaylistId(playlistId string) *model.PlaylistMetadata {
	var foundPlaylists []model.PlaylistMetadata
	database.GetConnection().Where(model.PlaylistMetadata{SpotifyPlaylistId: playlistId}).Find(&foundPlaylists)

	if len(foundPlaylists) > 0 {
		return &foundPlaylists[0]
	} else {
		return nil
	}
}
