package service

import (
	"github.com/47-11/spotifete/database"
	. "github.com/47-11/spotifete/model"
	"github.com/jinzhu/gorm"
	"github.com/zmb3/spotify"
	"golang.org/x/oauth2"
)

type SpotifyUserService struct {
	spotifyService SpotifyService
}

func (s SpotifyUserService) GetTotalUserCount() int {
	var count int
	database.Connection.Model(&SpotifyUser{}).Count(&count)
	return count
}

func (s SpotifyUserService) GetUserById(id uint) (SpotifyUser, error) {
	var users []SpotifyUser
	database.Connection.Where("id = ?", id).Find(&users)

	if len(users) == 1 {
		return users[0], nil
	} else {
		return SpotifyUser{}, EntryNotFoundError{Message: "User not found."}
	}
}

func (s SpotifyUserService) GetUserBySpotifyId(id string) (SpotifyUser, error) {
	var users []SpotifyUser
	database.Connection.Where("spotify_id = ?", id).Find(&users)

	if len(users) == 1 {
		return users[0], nil
	} else {
		return SpotifyUser{}, EntryNotFoundError{Message: "User not found."}
	}
}

func (s SpotifyUserService) GetOrCreateUserForToken(token *oauth2.Token) (*SpotifyUser, error) {
	client := s.spotifyService.GetAuthenticator().NewClient(token)
	user, err := client.CurrentUser()
	if err != nil {
		return nil, err
	}

	return s.GetOrCreateUserForSpotifyPrivateUser(user)
}

func (s SpotifyUserService) GetOrCreateUserForSpotifyPrivateUser(user *spotify.PrivateUser) (*SpotifyUser, error) {
	var internalUsers []SpotifyUser
	database.Connection.Where("spotify_id = ?", user.ID).Find(&internalUsers)

	if len(internalUsers) == 1 {
		return &internalUsers[0], nil
	} else {
		// No user found -> Create new
		newInternalUser := SpotifyUser{
			Model:     gorm.Model{},
			SpotifyId: user.ID,
		}

		database.Connection.NewRecord(newInternalUser)
		database.Connection.Create(&newInternalUser)

		return &newInternalUser, nil
	}
}
