package authentication

import (
	"github.com/47-11/spotifete/database"
	"github.com/47-11/spotifete/database/model"
)

func SetUserForSession(session model.LoginSession, user model.User) {
	session.UserId = &user.ID
	database.GetConnection().Save(session)
}
