package service

import (
	"github.com/47-11/spotifete/config"
	"github.com/gin-contrib/sessions"
	"github.com/zmb3/spotify"
	"golang.org/x/oauth2"
	"time"
)

type SpotifyService struct{}

var authenticator *spotify.Authenticator
var state string

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

func (s SpotifyService) GetState() string {
	if len(state) == 0 {
		state = "constant-for-now"
	}

	return state
}

func (s SpotifyService) NewAuthUrl() (string, string) {
	state := s.GetState()
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
