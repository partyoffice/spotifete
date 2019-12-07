package database

import "github.com/jinzhu/gorm"

type SongRequest struct {
	gorm.Model
	SessionId uint
	UserId    *uint
	TrackId   string
}
