package model

import (
	"github.com/47-11/spotifete/util"
	"github.com/zmb3/spotify"
	"gorm.io/gorm"
	"strings"
	"time"
)

type TrackMetadata struct {
	ID                     uint
	CreatedAt              time.Time
	UpdatedAt              time.Time
	DeletedAt              gorm.DeletedAt
	SpotifyTrackId         string `gorm:"primaryKey"`
	TrackName              string
	ArtistName             string
	AlbumName              string
	AlbumImageThumbnailUrl string
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
	trackMetadata.AlbumImageThumbnailUrl = util.FindSmallestImageUrlOrEmpty(spotifyTrack.Album.Images)

	return trackMetadata
}
