package dto

import (
	"time"
)

type ListeningSessionDto struct {
	Owner               SimpleUserDto        `json:"owner"`
	JoinId              string               `json:"joinId"`
	JoinIdHumanReadable string               `json:"joinIdHumanReadable"`
	Title               string               `json:"title"`
	CurrentlyPlaying    *TrackMetadataDto    `json:"currentlyPlaying"`
	UpNext              *TrackMetadataDto    `json:"upNext"`
	Queue               []TrackMetadataDto   `json:"queue"`
	QueueLastUpdated    time.Time            `json:"queueLastUpdated"`
	QueuePlaylistId     string               `json:"spotifyPlaylist"`
	FallbackPlaylist    *PlaylistMetadataDto `json:"fallbackPlaylist"`
}
