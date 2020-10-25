package user

import (
	"github.com/47-11/spotifete/database"
	"github.com/47-11/spotifete/database/model"
)

func FindSimpleUser(filter model.SimpleUser) *model.SimpleUser {
	users := FindSimpleUsers(filter)

	if len(users) == 1 {
		return &users[0]
	} else {
		return nil
	}
}

func FindSimpleUsers(filter model.SimpleUser) []model.SimpleUser {
	var users []model.SimpleUser
	database.GetConnection().Where(filter).Find(&users)
	return users
}

func FindFullUser(filter model.SimpleUser) *model.FullUser {
	users := FindFullUsers(filter)

	if len(users) == 1 {
		return &users[0]
	} else {
		return nil
	}
}

func FindFullUsers(filter model.SimpleUser) []model.FullUser {
	var users []model.FullUser
	database.GetConnection().Where(filter).Preload("ListeningSessions").Find(&users)
	return users
}
