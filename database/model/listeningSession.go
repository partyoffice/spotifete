package model

import "gorm.io/gorm"

type ListeningSession struct {
	gorm.Model
	Active           bool
	OwnerId          uint
	Owner            SimpleUser `gorm:"foreignKey:owner_id"`
	JoinId           *string
	QueuePlaylist    string
	Title            string
	FallbackPlaylist *string
}
