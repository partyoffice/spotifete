package model

import "github.com/jinzhu/gorm"

type SpotifyUser struct {
	gorm.Model
	SpotifyId string
}
