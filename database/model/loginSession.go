package model

import "gorm.io/gorm"

type LoginSession struct {
	gorm.Model
	SessionId        string
	UserId           *uint
	Active           bool
	CallbackRedirect string
}
