package model

import "github.com/jinzhu/gorm"

type SongRequest struct {
	gorm.Model
	SessionId   uint
	RequestedBy *uint
	SongId      string
}
