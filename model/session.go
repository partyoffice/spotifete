package model

import "github.com/jinzhu/gorm"

type Session struct {
	gorm.Model
	Active  bool
	OwnerId uint
}
