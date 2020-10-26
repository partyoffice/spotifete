package webapp

import "github.com/gin-gonic/gin"

func SetupStaticRouter(baseRouter *gin.Engine) {
	baseRouter.Static("/static/", "./resources/static/")
	baseRouter.StaticFile("favicon.ico", "./resources/favicon.ico")
}
