package model

import (
	"github.com/partyoffice/spotifete/shared"
	"github.com/zmb3/spotify"
)

type PlaylistMetadata struct {
	BaseModelWithoutId
	SpotifyPlaylistId string `gorm:"primaryKey" json:"spotify_playlist_id"`
	Name              string `json:"name"`
	TrackCount        uint   `json:"track_count"`
	ImageThumbnailUrl string `json:"image_thumbnail_url"`
	OwnerName         string `json:"owner_name"`
}

func (playlistMetadata PlaylistMetadata) FromFullPlaylist(fullPlaylist spotify.FullPlaylist) PlaylistMetadata {
	playlistMetadata.SpotifyPlaylistId = fullPlaylist.ID.String()
	playlistMetadata.Name = fullPlaylist.Name
	playlistMetadata.TrackCount = uint(fullPlaylist.Tracks.Total)
	playlistMetadata.ImageThumbnailUrl = shared.FindSmallestImageUrlOrEmpty(fullPlaylist.Images)
	playlistMetadata.OwnerName = fullPlaylist.Owner.DisplayName

	return playlistMetadata
}

func (playlistMetadata PlaylistMetadata) FromSimplePlaylist(simplePlaylist spotify.SimplePlaylist) PlaylistMetadata {
	playlistMetadata.SpotifyPlaylistId = simplePlaylist.ID.String()
	playlistMetadata.Name = simplePlaylist.Name
	playlistMetadata.TrackCount = uint(simplePlaylist.Tracks.Total)
	playlistMetadata.ImageThumbnailUrl = shared.FindSmallestImageUrlOrEmpty(simplePlaylist.Images)
	playlistMetadata.OwnerName = simplePlaylist.Owner.DisplayName

	return playlistMetadata
}
