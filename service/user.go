package service

import (
	"github.com/47-11/spotifete/database"
	. "github.com/47-11/spotifete/model/database"
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
	database.GetConnection().Model(&User{}).Count(&count)
	return count
}

func (userService) GetUserById(id uint) *User {
	var users []User
	database.GetConnection().Where(User{
		Model: gorm.Model{ID: id},
	}).Find(&users)

	if len(users) == 1 {
		return &users[0]
	} else {
		return nil
	}
}

func (userService) GetUserBySpotifyId(spotifyId string) *User {
	var users []User
	database.GetConnection().Where(User{SpotifyId: spotifyId}).Find(&users)

	if len(users) == 1 {
		return &users[0]
	} else {
		return nil
	}
}

func (userService) GetOrCreateUser(spotifyUser *spotify.PrivateUser) User {
	var users []User
	database.GetConnection().Where(User{SpotifyId: spotifyUser.ID}).Find(&users)

	if len(users) == 1 {
		return users[0]
	} else {
		// No user found -> Create new
		newUser := User{
			Model:              gorm.Model{},
			SpotifyId:          spotifyUser.ID,
			SpotifyDisplayName: spotifyUser.DisplayName,
		}

		database.GetConnection().NewRecord(newUser)
		database.GetConnection().Create(&newUser)

		return newUser
	}
}

func (userService) SetToken(user User, token oauth2.Token) {
	database.GetConnection().Model(&user).Updates(User{
		SpotifyAccessToken:  token.AccessToken,
		SpotifyRefreshToken: token.RefreshToken,
		SpotifyTokenType:    token.TokenType,
		// Save the expiry date as UTC because we don't save the timezone in the database and get the timetamp back as UTC when querying which confuses oauth2
		// TODO: Remove this once we figure out a way to properly handle timezones
		SpotifyTokenExpiry: token.Expiry.UTC(),
	})
}

func (s userService) CreateDto(user User, resolveAdditionalInformation bool) dto.UserDto {
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
