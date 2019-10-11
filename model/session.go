package model

import "github.com/jinzhu/gorm"

type Session struct {
	gorm.Model
	Active bool        `json:"active"`
	Owner  SpotifyUser `json:"owner"`
}
