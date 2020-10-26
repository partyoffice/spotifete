package api

import "github.com/gin-gonic/gin"

func SetupRouter(baseRouterGroup *gin.RouterGroup) {
	router := baseRouterGroup.Group("/user")

	router.GET("/me")
	router.GET("/id/:userId")
}
