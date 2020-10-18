package dto

import (
	"github.com/47-11/spotifete/database/model"
)

type TrackMetadataDto struct {
	SpotifyTrackId         string `json:"trackId"`
	TrackName              string `json:"trackName"`
	ArtistName             string `json:"artistName"`
	AlbumName              string `json:"albumName"`
	AlbumImageThumbnailUrl string `json:"albumImageThumbnailUrl"`
}

func (trackMetadataDto TrackMetadataDto) FromDatabaseModel(model model.TrackMetadata) TrackMetadataDto {
	trackMetadataDto.SpotifyTrackId = model.SpotifyTrackId
	trackMetadataDto.TrackName = model.TrackName
	trackMetadataDto.ArtistName = model.ArtistName
	trackMetadataDto.AlbumName = model.AlbumName
	trackMetadataDto.AlbumImageThumbnailUrl = model.AlbumImageThumbnailUrl

	return trackMetadataDto
}
