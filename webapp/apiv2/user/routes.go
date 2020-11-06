package user

import "github.com/gin-gonic/gin"

func SetupRoutes(baseRouterGroup *gin.RouterGroup) {
	router := baseRouterGroup.Group("/user")

	router.GET("/me", getCurrentUser)
	router.GET("/id/:userId")
}
