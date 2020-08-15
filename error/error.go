package error

import (
	. "github.com/47-11/spotifete/model/webapp/api/v1"
	"github.com/gin-gonic/gin"
	"net/http"
	"reflect"
)

type SpotifeteError interface {
	error
	StringResponse(c *gin.Context)
	JsonResponse(c *gin.Context)
	WithCause(cause error) SpotifeteError
	WithMessage(message string) SpotifeteError
	getHttpStatus() int
	shouldShowMessageToUser() bool
	shouldShowCauseToUser() bool
	getDefaultMessage() string
}

type BaseError struct {
	SpotifeteError
	Cause   *error
	Message *string
}

func (e BaseError) Error() (errorString string) {
	if e.shouldShowMessageToUser()  && e.Message != nil {
		errorString = *e.Message
	} else {
		errorString = e.getDefaultMessage()
	}

	if e.shouldShowCauseToUser() && e.Cause != nil{
		errorString += "\n"
		errorString += "caused by "
		errorString += reflect.TypeOf(e.Cause).Name()
		errorString += ": "
		errorString += (*e.Cause).Error()
	}

	return errorString
}

func (e BaseError) StringResponse(c *gin.Context) {
	c.String(e.getHttpStatus(), e.Error())
}

func (e BaseError) JsonResponse(c *gin.Context) {
	c.SecureJSON(e.getHttpStatus(), ErrorResponse{Message: e.Error()})
}

func (e BaseError) WithMessage(message string) SpotifeteError {
	e.Message = &message
	return e
}

func (e BaseError) WithCause(cause error) SpotifeteError {
	e.Cause = &cause
	return e
}

func (e BaseError) getHttpStatus() int {
	return http.StatusInternalServerError
}

func (e BaseError) shouldShowMessageToUser() bool {
	return true
}

func (e BaseError) shouldShowCauseToUser() bool {
	return true
}

func (e BaseError) getDefaultMessage() string {
	return "An error occured."
}
