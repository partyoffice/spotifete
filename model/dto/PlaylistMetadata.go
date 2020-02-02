package dto

type PlaylistMetadataDto struct {
	SpotifyPLaylistId         string `json:"trackId"`
	PLaylistName              string `json:"trackName"`
	TrackCount                uint   `json:"trackCount"`
	PlaylistImageThumbnailUrl string `json:"playlistImageThumbnailUrl"`
	CreatedByName             string `json:"createdBy"`
}
