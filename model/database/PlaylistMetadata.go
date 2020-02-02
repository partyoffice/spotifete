package database

import (
	"github.com/47-11/spotifete/util"
	"github.com/jinzhu/gorm"
	"github.com/zmb3/spotify"
)

type PlaylistMetadata struct {
	gorm.Model
	SpotifyPlaylistId         string
	PlaylistName              string
	TrackCount                uint
	PlaylistImageThumbnailUrl string
	CreatedByName             string
}

func (playlistMetadata PlaylistMetadata) SetMetadata(spotifyPlaylist spotify.FullPlaylist) PlaylistMetadata {
	playlistMetadata.SpotifyPlaylistId = spotifyPlaylist.ID.String()
	playlistMetadata.PlaylistName = spotifyPlaylist.Name
	playlistMetadata.TrackCount = uint(spotifyPlaylist.Tracks.Total)
	playlistMetadata.PlaylistImageThumbnailUrl = util.FindSmallestImage(spotifyPlaylist.Images).URL
	playlistMetadata.CreatedByName = spotifyPlaylist.Owner.DisplayName

	return playlistMetadata
}
