package database

import "github.com/jinzhu/gorm"

type LoginSession struct {
	gorm.Model
	SessionId string
	UserId    *uint
	Active    bool
}
