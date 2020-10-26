package error

import (
	"github.com/gin-gonic/gin"
)

type Response struct {
	Message string `json:"message"`
}

func (e SpotifeteError) SetJsonResponse(ctx *gin.Context) {
	ctx.JSON(e.HttpStatus, Response{Message: e.MessageForUser})
}

func (e SpotifeteError) SetStringResponse(ctx *gin.Context) {
	ctx.String(e.HttpStatus, e.MessageForUser)
}
