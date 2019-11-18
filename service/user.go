package service

import (
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
	database.Connection.Model(&User{}).Count(&count)
	return count
}

func (s UserService) GetUserById(id uint) (*User, error) {
	var users []User
	database.Connection.Where("id = ?", id).Find(&users)

	if len(users) == 1 {
		return &users[0], nil
	} else {
		return nil, EntryNotFoundError{Message: "User not found."}
	}
}

func (s UserService) GetUserBySpotifyId(id string) (*User, error) {
	var users []User
	database.Connection.Where("spotify_id = ?", id).Find(&users)

	if len(users) == 1 {
		return &users[0], nil
	} else {
		return nil, EntryNotFoundError{Message: "User not found."}
	}
}

func (s UserService) GetOrCreateUserForToken(token *oauth2.Token) (*User, error) {
	client := s.spotifyService.GetAuthenticator().NewClient(token)
	user, err := client.CurrentUser()
	if err != nil {
		return nil, err
	}

	return s.GetOrCreateUserForSpotifyPrivateUser(user)
}

func (s UserService) GetOrCreateUserForSpotifyPrivateUser(spotifyUser *spotify.PrivateUser) (*User, error) {
	var users []User
	database.Connection.Where("spotify_id = ?", spotifyUser.ID).Find(&users)

	if len(users) == 1 {
		return &users[0], nil
	} else {
		// No user found -> Create new
		newUser := User{
			Model:     gorm.Model{},
			SpotifyId: spotifyUser.ID,
		}

		database.Connection.NewRecord(newUser)
		database.Connection.Create(&newUser)

		return &newUser, nil
	}
}
