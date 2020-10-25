package api

import (
	. "github.com/47-11/spotifete/error"
	"github.com/47-11/spotifete/shared"
)

type NewSessionRequest struct {
	shared.AuthenticatedRequest
	ListeningSessionTitle string `json:"listening_session_title"`
}

func (r NewSessionRequest) Validate() *SpotifeteError {
	err := r.AuthenticatedRequest.Validate()
	if err != nil {
		return err
	}

	if r.ListeningSessionTitle == "" {
		return NewUserError("listening_session_title must not be empty")
	}

	return nil
}
