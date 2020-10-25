package shared

import . "github.com/47-11/spotifete/error"

type ErrorResponse struct {
	Message string `json:"message"`
}

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
