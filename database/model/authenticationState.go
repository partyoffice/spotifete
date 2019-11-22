package model

import "github.com/jinzhu/gorm"

type AuthenticationState struct {
	gorm.Model
	State  string
	Active bool
}
