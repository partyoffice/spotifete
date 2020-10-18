package authentication

import (
	"github.com/47-11/spotifete/config"
	"github.com/47-11/spotifete/database/model"
	"github.com/zmb3/spotify"
	"sync"
)

func authUrlForSession(session model.LoginSession) string {
	return getAuthenticator().AuthURL(session.SessionId)
}

func getAuthenticator() spotify.Authenticator {
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
