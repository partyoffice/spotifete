package users

import (
	"github.com/47-11/spotifete/authentication"
	"github.com/47-11/spotifete/database"
	"github.com/47-11/spotifete/database/model"
	. "github.com/47-11/spotifete/shared"
	"github.com/zmb3/spotify"
	"net/http"
)

var clientCache = map[uint]*spotify.Client{}

func Client(user model.SimpleUser) *spotify.Client {
	if client, ok := clientCache[user.ID]; ok {
		go refreshAndSaveTokenForUserIfNeccessary(*client, user)
		return client
	}

	token := user.GetToken()
	if token == nil {
		return nil
	}

	client := authentication.NewClientForToken(token)
	clientCache[user.ID] = &client
	go refreshAndSaveTokenForUserIfNeccessary(client, user)

	return &client
}

func refreshAndSaveTokenForUserIfNeccessary(client spotify.Client, user model.SimpleUser) *SpotifeteError {
	newToken, err := client.Token() // This should refresh the token if neccessary: https://github.com/zmb3/spotify/issues/108#issuecomment-568899119
	if err != nil {
		return NewError("Could not refresh Spotify access token. Please try to log out and log in again.", err, http.StatusUnauthorized)
	}

	if newToken.Expiry.After(user.SpotifyTokenExpiry) {
		// Token was updated, persist to database
		user = user.SetToken(newToken)

		// Do this in a goroutine so API calls don't have to wait for the database write to succeed
		go database.GetConnection().Save(&user)
	}

	return nil
}
