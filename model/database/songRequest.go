package database

import "github.com/jinzhu/gorm"

type SongRequest struct {
	gorm.Model
	SessionId uint
	UserId    *uint
	TrackId   string
	Status    SongRequestStatus
}

type SongRequestStatus string

const (
	PLAYED            SongRequestStatus = "PLAYED"
	CURRENTLY_PLAYING SongRequestStatus = "CURRENTLY_PLAYING"
	UP_NEXT           SongRequestStatus = "UP_NEXT"
	IN_QUEUE          SongRequestStatus = "IN_QUEUE"
)
