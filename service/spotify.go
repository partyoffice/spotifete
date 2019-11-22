package service

import (
	"errors"
	"github.com/47-11/spotifete/config"
	"github.com/47-11/spotifete/database"
	"github.com/47-11/spotifete/database/model"
	"github.com/gin-contrib/sessions"
	"github.com/jinzhu/gorm"
	"github.com/zmb3/spotify"
	"golang.org/x/oauth2"
	"math/rand"
	"time"
)

type SpotifyService struct{}

var authenticator *spotify.Authenticator

func (s SpotifyService) GetAuthenticator() spotify.Authenticator {
	if authenticator == nil {
		c := config.GetConfig()
		callbackUrl := c.GetString("server.baseUrl") + "/spotify/callback"

		newAuth := spotify.NewAuthenticator(callbackUrl, spotify.ScopePlaylistReadPrivate, spotify.ScopePlaylistModifyPrivate, spotify.ScopeUserLibraryRead, spotify.ScopeUserModifyPlaybackState, spotify.ScopeUserReadCurrentlyPlaying)
		newAuth.SetAuthInfo(c.GetString("spotify.id"), c.GetString("spotify.secret"))
		authenticator = &newAuth
	}

	return *authenticator
}

func (s SpotifyService) NewState() string {
	for {
		b := make([]rune, 256)
		for i := range b {
			b[i] = letterRunes[rand.Intn(len(letterRunes))]
		}
		newState := string(b)

		var existingStates []model.AuthenticationState
		database.Connection.Where("state = ?", newState).Find(&existingStates)
		if len(existingStates) == 0 {
			var newEntry = model.AuthenticationState{
				Model:  gorm.Model{},
				State:  newState,
				Active: true,
			}
			database.Connection.Create(newEntry)

			return newState
		}
	}
}

func (s SpotifyService) NewAuthUrl() (string, string) {
	state := s.NewState()
	database.Connection.Create(&model.AuthenticationState{
		Model:  gorm.Model{},
		State:  state,
		Active: true,
	})
	return s.GetAuthenticator().AuthURL(state), state
}

func (s SpotifyService) GetSpotifyTokenFromSession(session sessions.Session) (*oauth2.Token, error) {
	accessToken := session.Get("spotifyAccessToken")
	refreshToken := session.Get("spotifyRefreshToken")
	tokenExpiry := session.Get("spotifyTokenExpiry")
	tokenType := session.Get("spotifyTokenType")

	if accessToken != nil && refreshToken != nil && tokenExpiry != nil && tokenType != nil {
		tokenExpiryParsed, err := time.Parse(time.RFC3339, tokenExpiry.(string))
		if err != nil {
			return nil, err
		}

		return &oauth2.Token{
			AccessToken:  accessToken.(string),
			TokenType:    tokenType.(string),
			RefreshToken: refreshToken.(string),
			Expiry:       tokenExpiryParsed,
		}, nil
	} else {
		return nil, nil
	}
}

func (s SpotifyService) GetSpotifyClientUserFromSession(session sessions.Session) (*spotify.Client, error) {
	token, err := s.GetSpotifyTokenFromSession(session)
	if token == nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	client := s.GetAuthenticator().NewClient(token)
	return &client, nil
}

func (s SpotifyService) CheckTokenValidity(token *oauth2.Token) (bool, error) {
	client := s.GetAuthenticator().NewClient(token)
	user, err := client.CurrentUser()
	if err != nil && user == nil {
		// TODO actually verify that the token is invalid and not some other error occurred
		return false, err
	} else {
		return true, nil
	}
}

func (s SpotifyService) InvalidateState(state string) error {
	var entries []model.AuthenticationState
	database.Connection.Where("state = ?", state).Find(&entries)

	if len(entries) == 1 {
		entry := entries[0]
		if entry.Active {
			entry.Active = false
			database.Connection.Save(&entry)
			return nil
		} else {
			return errors.New("state has already been used")
		}
	} else {
		return errors.New("state not found")
	}
}
