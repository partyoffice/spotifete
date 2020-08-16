package error

import "net/http"

type userErrorBuilder struct {
	baseErrorBuilder
}

func (e userErrorBuilder) shouldShowMessageToUser() bool {
	return true
}

func (e userErrorBuilder) shouldShowCauseToUser() bool {
	return true
}


type Authentication struct {
	userErrorBuilder
}

func (e Authentication) getHttpStatus() int {
	return http.StatusUnauthorized
}


type IllegalArgument struct {
	userErrorBuilder
}

func (e IllegalArgument) getHttpStatus() int {
	return http.StatusBadRequest
}

func (e IllegalArgument) getDefaultMessage() string {
	return "Missing or invalid parameter(s)."
}

func (e IllegalArgument) shouldLogError() bool {
	return false
}


type IllegalState struct {
	userErrorBuilder
}

func (e IllegalState) getHttpStatus() int {
	return http.StatusInternalServerError
}
