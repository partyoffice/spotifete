package users

import (
	"fmt"
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
		go refreshAndSaveTokenForUserIfNecessary(client, user.ID)
		return client
	}

	token := user.GetToken()
	if token == nil {
		return nil
	}

	client := authentication.NewClientForToken(token)
	clientCache[user.ID] = &client
	go refreshAndSaveTokenForUserIfNecessary(&client, user.ID)

	return &client
}

func refreshAndSaveTokenForUserIfNecessary(client *spotify.Client, userId uint) *SpotifeteError {
	user := FindSimpleUser(model.SimpleUser{
		BaseModel: model.BaseModel{ID: userId},
	})
	if user == nil {
		return NewInternalError(fmt.Sprintf("Cannot refresh token for unknown user id %d", userId), nil)
	}

	newToken, err := client.Token() // This should refresh the token if necessary: https://github.com/zmb3/spotify/issues/108#issuecomment-568899119
	if err != nil {
		return NewError("Could not refresh Spotify access token. Please try to log out and log in again.", err, http.StatusUnauthorized)
	}

	if newToken.Expiry.After(user.SpotifyTokenExpiry) {
		updatedUser := user.SetToken(newToken)
		go database.GetConnection().Save(&updatedUser)
	}

	return nil
}
