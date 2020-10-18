package authentication

import (
	"github.com/47-11/spotifete/config"
	"github.com/47-11/spotifete/database/model"
	. "github.com/47-11/spotifete/error"
	"github.com/gin-gonic/gin"
	"github.com/zmb3/spotify"
	"golang.org/x/oauth2"
	"net/http"
	"sync"
)

func NewClientForToken(token *oauth2.Token) spotify.Client {
	return getSpotifyAuthenticator().NewClient(token)
}

func GetTokenFromCallback(callbackContext *gin.Context) (*oauth2.Token, *SpotifeteError) {
	state := callbackContext.Query("state")

	token, err := getSpotifyAuthenticator().Token(state, callbackContext.Request)
	if err != nil {
		return nil, NewError("Could not fetch access token from Spotify.", err, http.StatusUnauthorized)
	}

	return token, nil
}

func authUrlForSession(session model.LoginSession) string {
	return getSpotifyAuthenticator().AuthURL(session.SessionId)
}

func getSpotifyAuthenticator() spotify.Authenticator {
	createAuthenticatorOne.Do(createAuthenticator)
	return authenticator
}

func createAuthenticator() {
	c := config.Get()
	callbackUrl := c.SpotifeteConfiguration.BaseUrl + "/auth/callback"

	authenticator = spotify.NewAuthenticator(callbackUrl, spotify.ScopePlaylistReadPrivate, spotify.ScopePlaylistModifyPrivate, spotify.ScopeImageUpload, spotify.ScopeUserLibraryRead, spotify.ScopeUserModifyPlaybackState, spotify.ScopeUserReadCurrentlyPlaying, spotify.ScopeUserReadPrivate)
	authenticator.SetAuthInfo(c.SpotifyConfiguration.Id, c.SpotifyConfiguration.Secret)
}

var createAuthenticatorOne sync.Once
var authenticator spotify.Authenticator
