package controller

import "github.com/gin-gonic/gin"

type StaticController struct{ Controller }

func (StaticController) SetupWithBaseRouter(baseRouter *gin.Engine) {
	baseRouter.Static("/static/", "./resources/static/")
	baseRouter.StaticFile("favicon.ico", "./resources/favicon.ico")
}
