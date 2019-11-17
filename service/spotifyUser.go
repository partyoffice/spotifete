package service

import (
	"github.com/47-11/spotifete/database"
	. "github.com/47-11/spotifete/model"
)

type SpotifyUserService struct{}

func (s SpotifyUserService) GetTotalUserCount() int {
	var count int
	database.Connection.Model(&SpotifyUser{}).Count(&count)
	return count
}

func (s SpotifyUserService) GetUserById(id int64) (SpotifyUser, error) {
	var users []SpotifyUser
	database.Connection.Where("id = ?", id).Limit(1).Find(&users)

	if len(users) == 1 {
		return users[0], nil
	} else {
		return SpotifyUser{}, EntryNotFoundError{Message: "User not found."}
	}
}
