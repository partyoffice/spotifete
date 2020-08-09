package error

import (
	. "github.com/47-11/spotifete/model/webapp/api/v1"
	"github.com/gin-gonic/gin"
	"net/http"
)

type SpotifeteError interface {
	error
	StringResponse(c *gin.Context)
	JsonResponse(c *gin.Context)
	getHttpStatus() int
	shouldShowMessageToUser() bool
	getDefaultMessage() string
}

type BaseError struct {
	SpotifeteError
	Message string
}

func (e BaseError) Error() string {
	if e.shouldShowMessageToUser() {
		return e.Message
	} else {
		return e.getDefaultMessage()
	}
}

func (e BaseError) StringResponse(c *gin.Context) {
	c.String(e.getHttpStatus(), e.Error())
}

func (e BaseError) JsonResponse(c *gin.Context) {
	c.SecureJSON(e.getHttpStatus(), ErrorResponse{Message: e.Error()})
}

func (e BaseError) getHttpStatus() int {
	return http.StatusInternalServerError
}

func (e BaseError) shouldShowMessageToUser() bool {
	return false
}

func (e BaseError) getDefaultMessage() string {
	return "An error occured."
}
