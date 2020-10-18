package service

import (
	"github.com/47-11/spotifete/database"
	"github.com/47-11/spotifete/database/model"
	dto "github.com/47-11/spotifete/model/dto"
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
	database.GetConnection().Model(&model.User{}).Count(&count)
	return count
}

func (userService) GetUserById(id uint) *model.User {
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

func (userService) GetUserBySpotifyId(spotifyId string) *model.User {
	var users []model.User
	database.GetConnection().Where(model.User{SpotifyId: spotifyId}).Find(&users)

	if len(users) == 1 {
		return &users[0]
	} else {
		return nil
	}
}

func (userService) GetOrCreateUser(spotifyUser *spotify.PrivateUser) model.User {
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

func (userService) SetToken(user model.User, token oauth2.Token) {
	database.GetConnection().Model(&user).Updates(model.User{
		SpotifyAccessToken:  token.AccessToken,
		SpotifyRefreshToken: token.RefreshToken,
		SpotifyTokenType:    token.TokenType,
		SpotifyTokenExpiry:  token.Expiry,
	})
}

func (s userService) CreateDto(user model.User, resolveAdditionalInformation bool) dto.UserDto {
	result := dto.UserDto{}

	result.SpotifyId = user.SpotifyId
	result.SpotifyDisplayName = user.SpotifyDisplayName

	if resolveAdditionalInformation {
		result.ListeningSessions = []dto.ListeningSessionDto{}
		for _, session := range ListeningSessionService().GetActiveSessionsByOwnerId(user.ID) {
			result.ListeningSessions = append(result.ListeningSessions, ListeningSessionService().CreateDto(session, false))
		}
	}

	return result
}
