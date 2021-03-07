package model

type SongRequest struct {
	BaseModel
	SessionId      uint              `json:"session_id"`
	UserId         *uint             `json:"user_id"`
	SpotifyTrackId string            `json:"spotify_track_id"`
	TrackMetadata  TrackMetadata     `gorm:"foreignKey:spotify_track_id;references:spotify_track_id" json:"track_metadata"`
	Status         SongRequestStatus `json:"status"`
}

type SongRequestStatus string

const (
	StatusPlayed           SongRequestStatus = "PLAYED"
	StatusCurrentlyPlaying SongRequestStatus = "CURRENTLY_PLAYING"
	StatusUpNext           SongRequestStatus = "UP_NEXT"
	StatusInQueue          SongRequestStatus = "IN_QUEUE"
)
