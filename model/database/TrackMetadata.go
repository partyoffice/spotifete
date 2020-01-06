package database

import (
	"github.com/jinzhu/gorm"
	"github.com/zmb3/spotify"
	"strings"
)

type TrackMetadata struct {
	gorm.Model
	SpotifyTrackId         string
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

	// Find image with lowest quality
	smallestSize := -1
	for _, image := range spotifyTrack.Album.Images {
		currentImageSize := image.Width * image.Height
		if smallestSize < 0 || currentImageSize < smallestSize {
			smallestSize = currentImageSize
			trackMetadata.AlbumImageThumbnailUrl = image.URL
		}
	}

	return trackMetadata
}
