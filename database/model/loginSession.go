package model

type LoginSession struct {
	BaseModel
	SessionId        string
	UserId           *uint
	User             *SimpleUser `gorm:"foreignKey:user_id"`
	Active           bool
	CallbackRedirect string
}
