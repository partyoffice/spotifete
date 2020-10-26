package shared

import (
	"github.com/47-11/spotifete/authentication"
	. "github.com/47-11/spotifete/database/model"
	. "github.com/47-11/spotifete/shared"
)

type BaseRequest interface {
	Validate() *SpotifeteError
}

type AuthenticatedRequest struct {
	LoginSessionId string `json:"login_session_id"`
}

func (r AuthenticatedRequest) GetUser() (SimpleUser, *SpotifeteError) {
	if "" == r.LoginSessionId {
		return SimpleUser{}, NewUserError("Missing parameter login_session_id.")
	}

	session := authentication.GetSession(r.LoginSessionId)
	if session == nil {
		return SimpleUser{}, NewUserError("Unknown login_session_id.")
	}

	if !session.IsValid() {
		return SimpleUser{}, NewUserError("Invalid login session.")
	}

	if !session.IsAuthenticated() {
		return SimpleUser{}, NewUserError("Login session is not authenticated.")
	}

	if session.User == nil {
		return SimpleUser{}, NewUserError("No user found for login session.")
	}

	return *session.User, nil
}
