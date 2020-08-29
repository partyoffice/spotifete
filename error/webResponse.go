package error

import (
	"errors"
	v1 "github.com/47-11/spotifete/model/webapp/api/v1"
	"github.com/gin-gonic/gin"
	"github.com/google/logger"
	"net/http"
)

func SetStringResponseForContext(err error, c *gin.Context) {
	c.String(getHttpStatusForError(err), err.Error())
}

func SetJsonResponseForContext(err error, c *gin.Context) {
	c.JSON(getHttpStatusForError(err), v1.ErrorResponse{Message: err.Error()})
}

func getHttpStatusForError(err error) int {
	var spotifeteError spotifeteError

	if err == nil {
		logger.Warning("Nil error given. Using status code 500.")
		return http.StatusInternalServerError
	}
	if errors.As(err, &spotifeteError) {
		return spotifeteError.GetHttpStatus()
	} else {
		logger.Warning("Generic error given. Using status code 500.")
		return http.StatusInternalServerError
	}
}
