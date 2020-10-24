package model

import "gorm.io/gorm"

type ListeningSession struct {
	gorm.Model
	Active           bool
	OwnerId          uint
	JoinId           *string
	QueuePlaylist    string
	Title            string
	FallbackPlaylist *string
}
