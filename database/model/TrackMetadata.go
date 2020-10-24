package model

import (
	"github.com/47-11/spotifete/util"
	"gorm.io/gorm"
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
	trackMetadata.AlbumImageThumbnailUrl = util.FindSmallestImageUrlOrEmpty(spotifyTrack.Album.Images)

	return trackMetadata
}
