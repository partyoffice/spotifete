package listeningSession

import (
	"github.com/partyoffice/spotifete/database/model"
	"github.com/partyoffice/spotifete/users"
	"github.com/zmb3/spotify"
)

func Client(session model.FullListeningSession) *spotify.Client {
	return users.Client(session.Owner)
}
