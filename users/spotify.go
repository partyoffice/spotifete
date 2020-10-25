package users

import (
	"github.com/47-11/spotifete/database"
	"github.com/47-11/spotifete/database/model"
	"github.com/zmb3/spotify"
)

func getOrCreateFromSpotifyUser(spotifyUser *spotify.PrivateUser) model.SimpleUser {
	user := FindSimpleUser(model.SimpleUser{
		SpotifyId: spotifyUser.ID,
	})

	if user == nil {
		// No user found -> Create new
		user = &model.SimpleUser{
			SpotifyId:          spotifyUser.ID,
			SpotifyDisplayName: spotifyUser.DisplayName,
		}

		database.GetConnection().Create(user)

	}

	return *user
}
