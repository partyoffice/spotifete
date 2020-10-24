package model

import "gorm.io/gorm"

type SongRequest struct {
	gorm.Model
	SessionId      uint
	UserId         *uint
	SpotifyTrackId string
	Status         SongRequestStatus
}

type SongRequestStatus string

const (
	StatusPlayed           SongRequestStatus = "PLAYED"
	StatusCurrentlyPlaying SongRequestStatus = "CURRENTLY_PLAYING"
	StatusUpNext           SongRequestStatus = "UP_NEXT"
	StatusInQueue          SongRequestStatus = "IN_QUEUE"
)
