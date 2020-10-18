package user

import (
	"github.com/47-11/spotifete/database"
	"github.com/47-11/spotifete/database/model"
	dto "github.com/47-11/spotifete/model/dto"
	"github.com/jinzhu/gorm"
	"github.com/zmb3/spotify"
)

func GetTotalUserCount() int {
	var count int
	database.GetConnection().Model(&model.User{}).Count(&count)
	return count
}

func GetUserById(id uint) *model.User {
	var users []model.User
	database.GetConnection().Where(model.User{
		Model: gorm.Model{ID: id},
	}).Find(&users)

	if len(users) == 1 {
		return &users[0]
	} else {
		return nil
	}
}

func GetUserBySpotifyId(spotifyId string) *model.User {
	var users []model.User
	database.GetConnection().Where(model.User{SpotifyId: spotifyId}).Find(&users)

	if len(users) == 1 {
		return &users[0]
	} else {
		return nil
	}
}

func GetOrCreateUser(spotifyUser *spotify.PrivateUser) model.User {
	var users []model.User
	database.GetConnection().Where(model.User{SpotifyId: spotifyUser.ID}).Find(&users)

	if len(users) == 1 {
		return users[0]
	} else {
		// No user found -> Create new
		newUser := model.User{
			Model:              gorm.Model{},
			SpotifyId:          spotifyUser.ID,
			SpotifyDisplayName: spotifyUser.DisplayName,
		}

		database.GetConnection().NewRecord(newUser)
		database.GetConnection().Create(&newUser)

		return newUser
	}
}

func CreateDto(user model.User) dto.UserDto {
	result := dto.UserDto{}

	result.SpotifyId = user.SpotifyId
	result.SpotifyDisplayName = user.SpotifyDisplayName

	return result
}
