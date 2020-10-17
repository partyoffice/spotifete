package error

import (
	. "github.com/47-11/spotifete/model/webapp/api/v1"
	"github.com/gin-gonic/gin"
)

func (e SpotifeteError) SetJsonResponse(ctx *gin.Context) {
	ctx.JSON(e.HttpStatus, ErrorResponse{Message: e.MessageForUser})
}

func (e SpotifeteError) SetStringResponse(ctx *gin.Context) {
	ctx.String(e.HttpStatus, e.MessageForUser)
}
