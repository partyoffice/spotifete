package listeningSession

import (
	. "github.com/47-11/spotifete/error"
	"github.com/47-11/spotifete/webapp/apiv2/shared"
)

type NewSessionRequest struct {
	shared.AuthenticatedRequest
	ListeningSessionTitle string `json:"listening_session_title"`
}

func (r NewSessionRequest) Validate() *SpotifeteError {
	if "" == r.ListeningSessionTitle {
		return NewUserError("Missing parameter listening_session_title.")
	}

	return nil
}
