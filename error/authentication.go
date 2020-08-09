package error

import "net/http"

type AuthenticationError struct {
	BaseError
}

func (e AuthenticationError) getHttpStatus() int {
	return http.StatusUnauthorized
}

func (e AuthenticationError) shouldShowMessageToUser() bool {
	return true
}
