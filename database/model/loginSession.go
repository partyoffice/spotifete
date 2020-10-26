package model

import "time"

type LoginSession struct {
	BaseModel
	SessionId        string
	UserId           *uint
	User             *SimpleUser `gorm:"foreignKey:user_id"`
	Active           bool
	CallbackRedirect string
}

func (l LoginSession) IsAuthenticated() bool {
	return l.User != nil && l.IsValid()
}

func (l LoginSession) IsValid() bool {
	return l.Active && l.CreatedAt.AddDate(0, 0, 7).After(time.Now())
}
