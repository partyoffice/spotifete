package authentication

import (
	"github.com/47-11/spotifete/database"
	"github.com/47-11/spotifete/database/model"
	. "github.com/47-11/spotifete/error"
	"github.com/zmb3/spotify"
	"golang.org/x/oauth2"
	"net/http"
)

var clientCache = map[string]*spotify.Client{}

func GetClientForUser(user model.User) *spotify.Client {
	if client, ok := clientCache[user.SpotifyId]; ok {
		refreshAndSaveTokenForUserIfNeccessary(*client, user)
		return client
	}

	token := user.GetToken()
	if token == nil {
		return nil
	}

	client := NewClientForToken(token)
	refreshAndSaveTokenForUserIfNeccessary(client, user)
	clientCache[user.SpotifyId] = &client

	return &client
}

func refreshAndSaveTokenForUserIfNeccessary(client spotify.Client, user model.User) *SpotifeteError {
	newToken, err := client.Token() // This should refresh the token if neccessary: https://github.com/zmb3/spotify/issues/108#issuecomment-568899119
	if err != nil {
		return NewError("Could not refresh Spotify access token. Please try to log out and log in again.", err, http.StatusUnauthorized)
	}

	if newToken.Expiry.After(user.SpotifyTokenExpiry) {
		// Token was updated, persist to database
		// Do this in a goroutine so API calls don't have to wait for the database write to succeed
		go UpdateUserToken(user, *newToken)
	}

	return nil
}

func UpdateUserToken(user model.User, token oauth2.Token) {
	database.GetConnection().Model(&user).Updates(model.User{
		SpotifyAccessToken:  token.AccessToken,
		SpotifyRefreshToken: token.RefreshToken,
		SpotifyTokenType:    token.TokenType,
		SpotifyTokenExpiry:  token.Expiry,
	})
}

func SetUserForSession(session model.LoginSession, user model.User) {
	session.UserId = &user.ID
	database.GetConnection().Save(session)
}

func AddClientToCache(spotifyUser spotify.PrivateUser, client spotify.Client) {
	clientCache[spotifyUser.ID] = &client
}
