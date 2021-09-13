package shared

import (
	"fmt"

	"github.com/partyoffice/spotifete/authentication"
	. "github.com/partyoffice/spotifete/database/model"
	. "github.com/partyoffice/spotifete/shared"
	"github.com/partyoffice/spotifete/users"
)

type BaseRequest interface {
	Validate() *SpotifeteError
}

type AuthenticatedRequest struct {
	LoginSessionId string `json:"login_session_id"`
}

func (r AuthenticatedRequest) GetFullUser() (FullUser, *SpotifeteError) {
	simpleUser, spotifeteError := r.GetSimpleUser()
	if spotifeteError != nil {
		return FullUser{}, spotifeteError
	}

	fullUser := users.FindFullUser(SimpleUser{
		BaseModel: BaseModel{
			ID: simpleUser.ID,
		},
	})

	if fullUser == nil {
		return FullUser{}, NewInternalError(fmt.Sprintf("Could not find full user with ID %d", simpleUser.ID), nil)
	} else {
		return *fullUser, nil
	}
}

func (r AuthenticatedRequest) GetSimpleUser() (SimpleUser, *SpotifeteError) {
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
