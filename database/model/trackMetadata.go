package model

import (
	"github.com/partyoffice/spotifete/shared"
	"github.com/zmb3/spotify"
	"strings"
)

type TrackMetadata struct {
	BaseModelWithoutId
	SpotifyTrackId         string `gorm:"primaryKey" json:"spotify_track_id"`
	TrackName              string `json:"track_name"`
	ArtistName             string `json:"artist_name"`
	AlbumName              string `json:"album_name"`
	AlbumImageThumbnailUrl string `json:"album_image_thumbnail_url"`
}

func (trackMetadata TrackMetadata) SetMetadata(spotifyTrack spotify.FullTrack) TrackMetadata {
	trackMetadata.SpotifyTrackId = spotifyTrack.ID.String()
	trackMetadata.TrackName = spotifyTrack.Name
	trackMetadata.AlbumName = spotifyTrack.Album.Name

	var artistNames []string
	for _, artist := range spotifyTrack.Artists {
		artistNames = append(artistNames, artist.Name)
	}

	trackMetadata.ArtistName = strings.Join(artistNames, ", ")
	trackMetadata.AlbumImageThumbnailUrl = shared.FindSmallestImageUrlOrEmpty(spotifyTrack.Album.Images)

	return trackMetadata
}
