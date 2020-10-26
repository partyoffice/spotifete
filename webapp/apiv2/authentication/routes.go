package authentication

import (
	"github.com/gin-gonic/gin"
)

func SetupRoutes(baseRouterGroup *gin.RouterGroup) {
	router := baseRouterGroup.Group("/auth")

	router.GET("/session/new", newSession)
	router.GET("/session/id/:sessionId/is-authenticated", isSessionAuthenticated)
	router.PATCH("/session/id/:sessionId/invalidate", invalidateSession)
	router.Any("/success", callbackSuccess)
}
