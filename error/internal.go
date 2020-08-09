package error

import "net/http"

type InternalError struct {
	BaseError
}

func (e InternalError) getHttpStatus() int {
	return http.StatusInternalServerError
}

func (e InternalError) shouldShowMessageToUser() bool {
	return false
}
