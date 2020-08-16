package error

import "net/http"

type Internal struct {
	baseErrorBuilder
}

func (e Internal) getHttpStatus() int {
	return http.StatusInternalServerError
}

func (e Internal) shouldShowMessageToUser() bool {
	return false
}

func (e Internal) shouldShowCauseToUser() bool {
	return false
}

