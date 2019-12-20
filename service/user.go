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
	database.Connection.Model(&User{}).Count(&count)
	return count
}

func (userService) GetUserById(id uint) *User {
	var users []User
	database.Connection.Where(User{
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
	var sessions []ListeningSession
	database.Connection.Where(User{SpotifyId: spotifyId}).Find(&users).Related(&sessions)

	if len(users) == 1 {
		return &users[0]
	} else {
		return nil
	}
}

func (userService) GetOrCreateUser(spotifyUser *spotify.PrivateUser) *User {
	var users []User
	database.Connection.Where(User{SpotifyId: spotifyUser.ID}).Find(&users)

	if len(users) == 1 {
		return &users[0]
	} else {
		// No user found -> Create new
		newUser := User{
			Model:              gorm.Model{},
			SpotifyId:          spotifyUser.ID,
			SpotifyDisplayName: spotifyUser.DisplayName,
		}

		database.Connection.NewRecord(newUser)
		database.Connection.Create(&newUser)

		return &newUser
	}
}

func (userService) SetToken(user *User, token *oauth2.Token) {
	user.SpotifyAccessToken = token.AccessToken
	user.SpotifyRefreshToken = token.RefreshToken
	user.SpotifyTokenType = token.TokenType
	user.SpotifyTokenExpiry = token.Expiry

	database.Connection.Save(user)
}

func (s userService) CreateDto(user User, resolveAdditionalInformation bool) dto.UserDto {
	result := dto.UserDto{}

	result.SpotifyId = user.SpotifyId
	result.SpotifyDisplayName = user.SpotifyDisplayName

	result.ListeningSessions = []dto.ListeningSessionDto{}
	for _, session := range ListeningSessionService().GetActiveSessionsByOwnerId(user.ID) {
		result.ListeningSessions = append(result.ListeningSessions, ListeningSessionService().CreateDto(session, resolveAdditionalInformation))
	}

	return result
}
