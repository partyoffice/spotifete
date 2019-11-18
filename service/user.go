package service

import (
	"github.com/47-11/SpotiFete/database/model"
	"github.com/47-11/spotifete/database"
	. "github.com/47-11/spotifete/model"
	"github.com/jinzhu/gorm"
	"github.com/zmb3/spotify"
	"golang.org/x/oauth2"
)

type UserService struct {
	spotifyService SpotifyService
}

func (s UserService) GetTotalUserCount() int {
	var count int
	database.Connection.Model(&model.User{}).Count(&count)
	return count
}

func (s UserService) GetUserById(id uint) (*model.User, error) {
	var users []model.User
	database.Connection.Where("id = ?", id).Find(&users)

	if len(users) == 1 {
		return &users[0], nil
	} else {
		return nil, EntryNotFoundError{Message: "User not found."}
	}
}

func (s UserService) GetUserBySpotifyId(id string) (*model.User, error) {
	var users []model.User
	database.Connection.Where("spotify_id = ?", id).Find(&users)

	if len(users) == 1 {
		return &users[0], nil
	} else {
		return nil, EntryNotFoundError{Message: "User not found."}
	}
}

func (s UserService) GetOrCreateUserForToken(token *oauth2.Token) (*model.User, error) {
	client := s.spotifyService.GetAuthenticator().NewClient(token)
	user, err := client.CurrentUser()
	if err != nil {
		return nil, err
	}

	return s.GetOrCreateUserForSpotifyPrivateUser(user)
}

func (s UserService) GetOrCreateUserForSpotifyPrivateUser(spotifyUser *spotify.PrivateUser) (*model.User, error) {
	var users []model.User
	database.Connection.Where("spotify_id = ?", spotifyUser.ID).Find(&users)

	if len(users) == 1 {
		return &users[0], nil
	} else {
		// No user found -> Create new
		newUser := model.User{
			Model:     gorm.Model{},
			SpotifyId: spotifyUser.ID,
		}

		database.Connection.NewRecord(newUser)
		database.Connection.Create(&newUser)

		return &newUser, nil
	}
}
