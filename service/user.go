package service

import (
	"github.com/47-11/spotifete/database"
	"github.com/47-11/spotifete/database/model"
	"github.com/47-11/spotifete/webapp/model/api/v1/dto"
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

func (userService) GetUserById(id uint) *model.User {
	var users []model.User
	database.Connection.Where(model.User{
		Model: gorm.Model{ID: id},
	}).Find(&users)

	if len(users) == 1 {
		return &users[0]
	} else {
		return nil
	}
}

func (userService) GetUserBySpotifyId(spotifyId string) *model.User {
	var users []model.User
	var sessions []model.ListeningSession
	database.Connection.Where(model.User{SpotifyId: spotifyId}).Find(&users).Related(&sessions)

	if len(users) == 1 {
		return &users[0]
	} else {
		return nil
	}
}

func (userService) GetOrCreateUser(spotifyUser *spotify.PrivateUser) *model.User {
	var users []model.User
	database.Connection.Where(model.User{SpotifyId: spotifyUser.ID}).Find(&users)

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

func (s userService) CreateDtoWithAdditionalInformation(user *model.User) dto.UserDto {
	result := dto.UserDto{}.FromDatabaseModel(user)

	for _, session := range ListeningSessionService().GetActiveSessionsByOwnerId(user.ID) {
		result.ListeningSessions = append(result.ListeningSessions, dto.ListeningSessionDto{}.FromDatabaseModel(session))
	}

	return result
}
