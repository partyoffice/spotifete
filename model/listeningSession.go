package model

import "github.com/jinzhu/gorm"

type ListeningSession struct {
	gorm.Model
	Active  bool
	OwnerId uint
	JoinId  string
}
