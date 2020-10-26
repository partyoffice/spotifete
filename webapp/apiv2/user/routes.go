package user

import "github.com/gin-gonic/gin"

func SetupRoutes(baseRouterGroup *gin.RouterGroup) {
	router := baseRouterGroup.Group("/user")

	router.GET("/me")
	router.GET("/id/:userId")
}
