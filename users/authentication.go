package users

import (
	"net/http"

	"github.com/partyoffice/spotifete/authentication"
	"github.com/partyoffice/spotifete/database"
	"github.com/partyoffice/spotifete/database/model"
	. "github.com/partyoffice/spotifete/shared"
	"github.com/zmb3/spotify"
	"golang.org/x/oauth2"
)

func CreateAuthenticatedUser(token *oauth2.Token, loginSession model.LoginSession) (model.SimpleUser, *SpotifeteError) {
	client := authentication.NewClientForToken(token)
	spotifyUser, err := client.CurrentUser()
	if err != nil {
		return model.SimpleUser{}, NewError("Could not get user information from Spotify.", err, http.StatusInternalServerError)
	}

	persistedUser := getOrCreateFromSpotifyUser(spotifyUser)
	persistedUser = persistedUser.SetToken(token)
	database.GetConnection().Save(persistedUser)

	loginSession.UserId = &persistedUser.ID
	database.GetConnection().Save(&loginSession)

	return persistedUser, nil
}

func getOrCreateFromSpotifyUser(spotifyUser *spotify.PrivateUser) model.SimpleUser {
	user := FindSimpleUser(model.SimpleUser{
		SpotifyId: spotifyUser.ID,
	})

	if user == nil {
		// No user found -> Create new
		user = &model.SimpleUser{
			SpotifyId:          spotifyUser.ID,
			SpotifyDisplayName: spotifyUser.DisplayName,
			Country:            spotifyUser.Country,
		}

		database.GetConnection().Create(user)

	}

	return *user
}
