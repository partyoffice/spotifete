package service

import (
	"github.com/47-11/spotifete/database"
	"github.com/47-11/spotifete/database/model"
	. "github.com/47-11/spotifete/model"
	"github.com/jinzhu/gorm"
	"github.com/zmb3/spotify"
	"golang.org/x/oauth2"
	"sync"
)

type userService struct{}

var userServiceInstance *userService
var userServiceOnce sync.Once

func UserService() *userService {
	userServiceOnce.Do(func() {
		userServiceInstance = &userService{}
	})
	return userServiceInstance
}

func (userService) GetTotalUserCount() int {
	var count int
	database.Connection.Model(&model.User{}).Count(&count)
	return count
}

func (userService) GetUserById(id uint) (*model.User, error) {
	var users []model.User
	database.Connection.Where("id = ?", id).Find(&users)

	if len(users) == 1 {
		return &users[0], nil
	} else {
		return nil, EntryNotFoundError{Message: "User not found."}
	}
}

func (userService) GetUserBySpotifyId(id string) (*model.User, error) {
	var users []model.User
	database.Connection.Where("spotify_id = ?", id).Find(&users)

	if len(users) == 1 {
		return &users[0], nil
	} else {
		return nil, EntryNotFoundError{Message: "User not found."}
	}
}

func (userService) GetOrCreateUser(spotifyUser *spotify.PrivateUser) *model.User {
	var users []model.User
	database.Connection.Where("spotify_id = ?", spotifyUser.ID).Find(&users)

	if len(users) == 1 {
		return &users[0]
	} else {
		// No user found -> Create new
		newUser := model.User{
			Model:              gorm.Model{},
			SpotifyId:          spotifyUser.ID,
			SpotifyDisplayName: spotifyUser.DisplayName,
		}

		database.Connection.NewRecord(newUser)
		database.Connection.Create(&newUser)

		return &newUser
	}
}

func (userService) SetToken(user *model.User, token *oauth2.Token) {
	user.SpotifyAccessToken = token.AccessToken
	user.SpotifyRefreshToken = token.RefreshToken
	user.SpotifyTokenType = token.TokenType
	user.SpotifyTokenExpiry = token.Expiry

	database.Connection.Save(user)
}
