package listeningSession

import (
	"github.com/47-11/spotifete/database"
	"github.com/47-11/spotifete/database/model"
	"github.com/zmb3/spotify"
)

func AddOrUpdatePlaylistMetadata(playlist spotify.FullPlaylist) model.PlaylistMetadata {
	knownPlaylistMetadata := GetPlaylistMetadataBySpotifyPlaylistId(playlist.ID.String())
	if knownPlaylistMetadata != nil {
		updatedPlaylistMetadata := knownPlaylistMetadata.FromFullPlaylist(playlist)

		database.GetConnection().Save(&updatedPlaylistMetadata)

		return updatedPlaylistMetadata
	} else {
		newPlaylistMetadata := model.PlaylistMetadata{}.FromFullPlaylist(playlist)

		database.GetConnection().Create(&newPlaylistMetadata)

		return newPlaylistMetadata
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
