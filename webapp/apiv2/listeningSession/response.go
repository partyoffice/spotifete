package listeningSession

import (
	"github.com/47-11/spotifete/database/model"
	"time"
)

type SearchTracksResponse struct {
	Query  string                `json:"query"`
	Tracks []model.TrackMetadata `json:"tracks"`
}

type SearchPlaylistResponse struct {
	Query     string                   `json:"query"`
	Playlists []model.PlaylistMetadata `json:"playlists"`
}

type GetSessionQueueResponse struct {
	CurrentlyPlayingRequest *model.SongRequest  `json:"currently_playing"`
	UpNextRequest           *model.SongRequest  `json:"up_next"`
	Queue                   []model.SongRequest `json:"queue"`
}

type QueueLastUpdatedResponse struct {
	QueueLastUpdated time.Time `json:"queue_last_updated"`
}
