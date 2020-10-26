package listeningSession

import (
	"github.com/gin-gonic/gin"
)

func SetupRoutes(baseRouterGroup *gin.RouterGroup) {
	router := baseRouterGroup.Group("/session")

	router.POST("/new", newSession)
	router.GET("/id/:joinId", getSession)
	router.DELETE("/id/:joinId", closeSession)
	router.GET("/id/:joinId/search/track", searchTrack)
	router.GET("/id/:joinId/search/playlist", searchPlaylist)
	router.POST("/id/:joinId/request-track")
	router.GET("/id/:joinId/queue-last-updated")
	router.GET("/id/:joinId/qrcode")
}
