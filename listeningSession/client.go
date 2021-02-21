package listeningSession

import (
	"github.com/47-11/spotifete/database/model"
	"github.com/47-11/spotifete/users"
	"github.com/zmb3/spotify"
)

func Client(session model.FullListeningSession) *spotify.Client {
	return users.Client(session.Owner)
}
