package controller

import "github.com/gin-gonic/gin"

type Controller interface {
	SetupWithBaseRouter(baseRouter *gin.Engine)
}
