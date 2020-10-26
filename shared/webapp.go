package shared

import . "github.com/47-11/spotifete/error"

type BaseRequest interface {
	Validate() *SpotifeteError
}

type AuthenticatedRequest struct {
	BaseRequest
	LoginSessionId string `json:"login_session_id"`
}

func (r AuthenticatedRequest) Validate() *SpotifeteError {
	if r.LoginSessionId == "" {
		return NewUserError("login_session_id must not be empty")
	}

	return nil
}

// TODO: Wäre ganz nice wenn man auf einer AuthenticatedRequest direkt sagen könnte GetLoginSession() und / oder direkt den user kriegen könnte.
// Das würde ne menge doppelten code sparen
// Geht aber nicht hier weil das dann ein zyklischer Import wäre. Macht vielleicht im authentication package Sinn?
