package database

import (
	"github.com/jinzhu/gorm"
	"github.com/zmb3/spotify"
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
	trackMetadata.TrackName = spotifyTrack.Name
	trackMetadata.AlbumName = spotifyTrack.Album.Name

	trackMetadata.ArtistName = spotifyTrack.Artists[0].Name // TODO: add all artists

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
