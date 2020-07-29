package webapp

import "github.com/gin-gonic/gin"

type StaticController struct{}

func (StaticController) SetupRoutes(baseRouter *gin.Engine) {
	baseRouter.Static("/static/", "./resources/static/")
	baseRouter.StaticFile("favicon.ico", "./resources/favicon.ico")
}
