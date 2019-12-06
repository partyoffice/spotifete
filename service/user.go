package service

import (
	"github.com/47-11/spotifete/database"
	database2 "github.com/47-11/spotifete/model/database"
	dto2 "github.com/47-11/spotifete/model/dto"
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
	database.Connection.Model(&database2.User{}).Count(&count)
	return count
}

func (userService) GetUserById(id uint) *database2.User {
	var users []database2.User
	database.Connection.Where(database2.User{
		Model: gorm.Model{ID: id},
	}).Find(&users)

	if len(users) == 1 {
		return &users[0]
	} else {
		return nil
	}
}

func (userService) GetUserBySpotifyId(spotifyId string) *database2.User {
	var users []database2.User
	var sessions []database2.ListeningSession
	database.Connection.Where(database2.User{SpotifyId: spotifyId}).Find(&users).Related(&sessions)

	if len(users) == 1 {
		return &users[0]
	} else {
		return nil
	}
}

func (userService) GetOrCreateUser(spotifyUser *spotify.PrivateUser) *database2.User {
	var users []database2.User
	database.Connection.Where(database2.User{SpotifyId: spotifyUser.ID}).Find(&users)

	if len(users) == 1 {
		return &users[0]
	} else {
		// No user found -> Create new
		newUser := database2.User{
			Model:              gorm.Model{},
			SpotifyId:          spotifyUser.ID,
			SpotifyDisplayName: spotifyUser.DisplayName,
		}

		database.Connection.NewRecord(newUser)
		database.Connection.Create(&newUser)

		return &newUser
	}
}

func (userService) SetToken(user *database2.User, token *oauth2.Token) {
	user.SpotifyAccessToken = token.AccessToken
	user.SpotifyRefreshToken = token.RefreshToken
	user.SpotifyTokenType = token.TokenType
	user.SpotifyTokenExpiry = token.Expiry

	database.Connection.Save(user)
}

func (s userService) CreateDtoWithAdditionalInformation(user *database2.User) dto2.UserDto {
	result := dto2.UserDto{}.FromDatabaseModel(user)

	for _, session := range ListeningSessionService().GetActiveSessionsByOwnerId(user.ID) {
		result.ListeningSessions = append(result.ListeningSessions, dto2.ListeningSessionDto{}.FromDatabaseModel(session))
	}

	return result
}
