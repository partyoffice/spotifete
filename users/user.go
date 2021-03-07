package users

import (
	"fmt"
	"github.com/47-11/spotifete/database"
	"github.com/47-11/spotifete/database/model"
	. "github.com/47-11/spotifete/shared"
)

func FindSimpleUser(filter model.SimpleUser) *model.SimpleUser {
	users := FindSimpleUsers(filter)

	resultCount := len(users)
	if resultCount == 1 {
		return &users[0]
	} else if resultCount == 0 {
		return nil
	} else {
		NewInternalError(fmt.Sprintf("Got more than one result for filter %v", filter), nil)
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

	resultCount := len(users)
	if resultCount == 1 {
		return &users[0]
	} else if resultCount == 0 {
		return nil
	} else {
		NewInternalError(fmt.Sprintf("Got more than one result for filter %v", filter), nil)
		return nil
	}
}

func FindFullUsers(filter model.SimpleUser) []model.FullUser {
	var users []model.FullUser
	database.GetConnection().Where(filter).Preload("ListeningSessions", "active = true").Find(&users)
	return users
}
