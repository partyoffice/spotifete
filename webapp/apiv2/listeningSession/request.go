package listeningSession

import (
	. "github.com/47-11/spotifete/shared"
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

type RequestTrackRequest struct {
	TrackId string `json:"track_id"`
}

func (r RequestTrackRequest) Validate() *SpotifeteError {
	if "" == r.TrackId {
		return NewUserError("Missing parameter track_id.")
	}

	return nil
}
