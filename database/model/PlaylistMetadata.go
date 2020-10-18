package model

import (
	"github.com/47-11/spotifete/util"
	"github.com/jinzhu/gorm"
	"github.com/zmb3/spotify"
)

type PlaylistMetadata struct {
	gorm.Model
	SpotifyPlaylistId string
	Name              string
	TrackCount        uint
	ImageThumbnailUrl string
	OwnerName         string
}

func (playlistMetadata PlaylistMetadata) FromFullPlaylist(fullPlaylist spotify.FullPlaylist) PlaylistMetadata {
	playlistMetadata.SpotifyPlaylistId = fullPlaylist.ID.String()
	playlistMetadata.Name = fullPlaylist.Name
	playlistMetadata.TrackCount = uint(fullPlaylist.Tracks.Total)
	playlistMetadata.ImageThumbnailUrl = util.FindSmallestImage(fullPlaylist.Images).URL
	playlistMetadata.OwnerName = fullPlaylist.Owner.DisplayName

	return playlistMetadata
}

func (playlistMetadata PlaylistMetadata) FromSimplePlaylist(simplePlaylist spotify.SimplePlaylist) PlaylistMetadata {
	playlistMetadata.SpotifyPlaylistId = simplePlaylist.ID.String()
	playlistMetadata.Name = simplePlaylist.Name
	playlistMetadata.TrackCount = uint(simplePlaylist.Tracks.Total)
	playlistMetadata.ImageThumbnailUrl = util.FindSmallestImage(simplePlaylist.Images).URL
	playlistMetadata.OwnerName = simplePlaylist.Owner.DisplayName

	return playlistMetadata
}
