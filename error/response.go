package error

import (
	"github.com/47-11/spotifete/shared"
	"github.com/gin-gonic/gin"
)

func (e SpotifeteError) SetJsonResponse(ctx *gin.Context) {
	ctx.JSON(e.HttpStatus, shared.ErrorResponse{Message: e.MessageForUser})
}

func (e SpotifeteError) SetStringResponse(ctx *gin.Context) {
	ctx.String(e.HttpStatus, e.MessageForUser)
}
