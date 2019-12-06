package dto

type SearchTracksResultDto struct {
	TrackId       string `json:"trackId"`
	TrackName     string `json:"trackName"`
	ArtistName    string `json:"artistName"`
	AlbumName     string `json:"albumName"`
	AlbumImageUrl string `json:"albumImageUrl"`
}
