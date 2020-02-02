package dto

import . "github.com/47-11/spotifete/model/database"

type PlaylistMetadataDto struct {
	SpotifyPLaylistId         string `json:"trackId"`
	PlaylistName              string `json:"trackName"`
	TrackCount                uint   `json:"trackCount"`
	PlaylistImageThumbnailUrl string `json:"playlistImageThumbnailUrl"`
	CreatedByName             string `json:"createdBy"`
}

func (playlistMetadataDto PlaylistMetadataDto) FromDatabaseModel(model PlaylistMetadata) PlaylistMetadataDto {
	playlistMetadataDto.SpotifyPLaylistId = model.SpotifyPlaylistId
	playlistMetadataDto.PlaylistName = model.PlaylistName
	playlistMetadataDto.TrackCount = model.TrackCount
	playlistMetadataDto.PlaylistImageThumbnailUrl = model.PlaylistImageThumbnailUrl
	playlistMetadataDto.CreatedByName = model.CreatedByName

	return playlistMetadataDto
}
