package service

import (
	"github.com/47-11/spotifete/config"
	"github.com/47-11/spotifete/database"
	"github.com/47-11/spotifete/database/model"
	"github.com/gin-contrib/sessions"
	"github.com/jinzhu/gorm"
	"github.com/zmb3/spotify"
	"golang.org/x/oauth2"
	"sync"
	"time"
)

type spotifyService struct {
	authenticator *spotify.Authenticator
}

var spotifyServiceInstance *spotifyService
var spotifyServiceOnce sync.Once

func SpotifyService() *spotifyService {
	spotifyServiceOnce.Do(func() {
		c := config.GetConfig()
		callbackUrl := c.GetString("server.baseUrl") + "/spotify/callback"

		newAuth := spotify.NewAuthenticator(callbackUrl, spotify.ScopePlaylistReadPrivate, spotify.ScopePlaylistModifyPrivate, spotify.ScopeUserLibraryRead, spotify.ScopeUserModifyPlaybackState, spotify.ScopeUserReadCurrentlyPlaying)
		newAuth.SetAuthInfo(c.GetString("spotify.id"), c.GetString("spotify.secret"))

		spotifyServiceInstance = &spotifyService{
			authenticator: &newAuth,
		}
	})
	return spotifyServiceInstance
}

func (s spotifyService) GetAuthenticator() spotify.Authenticator {
	return *s.authenticator
}

func (s spotifyService) NewAuthUrl() (string, string) {
	sessionId := LoginSessionService().newSessionId()
	database.Connection.Create(&model.LoginSession{
		Model:     gorm.Model{},
		SessionId: sessionId,
		UserId:    nil,
		Active:    true,
	})
	return s.GetAuthenticator().AuthURL(sessionId), sessionId
}

func (spotifyService) GetSpotifyTokenFromSession(session sessions.Session) (*oauth2.Token, error) {
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

func (s spotifyService) GetSpotifyClientUserFromSession(session sessions.Session) (*spotify.Client, error) {
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

func (s spotifyService) CheckTokenValidity(token *oauth2.Token) (bool, error) {
	client := s.GetAuthenticator().NewClient(token)
	user, err := client.CurrentUser()
	if err != nil && user == nil {
		// TODO actually verify that the token is invalid and not some other error occurred
		return false, err
	} else {
		return true, nil
	}
}

func (s spotifyService) SearchTrack(client *spotify.Client, query string) (*spotify.SearchResult, error) {
	result, err := client.Search(query, spotify.SearchTypeTrack)
	if err != nil {
		return nil, err
	}

	// TODO: Refine search results

	return result, nil
}
