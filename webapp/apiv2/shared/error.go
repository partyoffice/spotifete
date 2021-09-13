package shared

import (
	"github.com/gin-gonic/gin"
	. "github.com/partyoffice/spotifete/shared"
)

type ErrorResponse struct {
	Message string `json:"message"`
}

func SetJsonError(error SpotifeteError, ctx *gin.Context) {
	ctx.JSON(error.HttpStatus, ErrorResponse{Message: error.MessageForUser})
}
