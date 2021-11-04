package model

type SongRequest struct {
	BaseModel
	SessionId      uint          `json:"session_id"`
	SpotifyTrackId string        `json:"spotify_track_id"`
	TrackMetadata  TrackMetadata `gorm:"foreignKey:spotify_track_id;references:spotify_track_id" json:"track_metadata"`
	Played         bool          `json:"-"`
	Locked         bool          `json:"-"`
	RequestedBy    string        `json:"requested_by"`
	Weight         int64         `json:"-"`
}
