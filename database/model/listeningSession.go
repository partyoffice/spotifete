package model

type SimpleListeningSession struct {
	BaseModel
	Active             bool
	OwnerId            uint
	JoinId             *string
	QueuePlaylistId    string `gorm:"column:queue_playlist"`
	Title              string
	FallbackPlaylistId *string `gorm:"column:fallback_playlist"`
}

func (SimpleListeningSession) TableName() string {
	return "listening_sessions"
}

type FullListeningSession struct {
	SimpleListeningSession
	Owner                    SimpleUser        `gorm:"foreignKey:owner_id"`
	QueuePlaylistMetadata    PlaylistMetadata  `gorm:"foreignKey:queue_playlist"`
	FallbackPlaylistMetadata *PlaylistMetadata `gorm:"foreignKey:fallback_playlist"`
}

func (FullListeningSession) TableName() string {
	return "listening_sessions"
}
