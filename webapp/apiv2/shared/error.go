package shared

import (
	. "github.com/47-11/spotifete/shared"
	"github.com/gin-gonic/gin"
)

type ErrorResponse struct {
	Message string `json:"message"`
}

func SetJsonError(error SpotifeteError, ctx *gin.Context) {
	ctx.JSON(error.HttpStatus, ErrorResponse{Message: error.MessageForUser})
}
