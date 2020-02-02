package dto

import . "github.com/47-11/spotifete/model/database"

type PlaylistMetadataDto struct {
	SpotifyPLaylistId string `json:"spotifyPlaylistId"`
	Name              string `json:"name"`
	TrackCount        uint   `json:"trackCount"`
	ImageThumbnailUrl string `json:"imageThumbnailUrl"`
	OwnerName         string `json:"createdBy"`
}

func (playlistMetadataDto PlaylistMetadataDto) FromDatabaseModel(model PlaylistMetadata) PlaylistMetadataDto {
	playlistMetadataDto.SpotifyPLaylistId = model.SpotifyPlaylistId
	playlistMetadataDto.Name = model.Name
	playlistMetadataDto.TrackCount = model.TrackCount
	playlistMetadataDto.ImageThumbnailUrl = model.ImageThumbnailUrl
	playlistMetadataDto.OwnerName = model.OwnerName

	return playlistMetadataDto
}
