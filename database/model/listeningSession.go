package model

type SimpleListeningSession struct {
	BaseModel
	Active             bool    `json:"active"`
	OwnerId            uint    `json:"owner_id"`
	JoinId             string  `json:"join_id"`
	QueuePlaylistId    string  `gorm:"column:queue_playlist" json:"queue_playlist_id"`
	Title              string  `json:"title"`
	FallbackPlaylistId *string `gorm:"column:fallback_playlist" json:"fallback_playlist_id"`
}

func (SimpleListeningSession) TableName() string {
	return "listening_sessions"
}

type FullListeningSession struct {
	SimpleListeningSession
	Owner                    SimpleUser        `gorm:"foreignKey:owner_id" json:"owner"`
	FallbackPlaylistMetadata *PlaylistMetadata `gorm:"foreignKey:fallback_playlist;references:spotify_playlist_id" json:"fallback_playlist_metadata"`
}

func (FullListeningSession) TableName() string {
	return "listening_sessions"
}
