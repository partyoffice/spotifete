package listeningSession

import (
	"github.com/gin-gonic/gin"
)

func SetupRoutes(baseRouterGroup *gin.RouterGroup) {
	router := baseRouterGroup.Group("/session")

	router.POST("/new", newSession)
	router.GET("/id/:joinId", getSession)
	router.DELETE("/id/:joinId", closeSession)
	router.GET("/id/:joinId/queue", getSessionQueue)
	router.GET("/id/:joinId/queue/last-updated", queueLastUpdated)
	router.GET("/id/:joinId/qrcode", qrCode)
	router.GET("/id/:joinId/search/track", searchTrack)
	router.GET("/id/:joinId/search/playlist", searchPlaylist)
	router.POST("/id/:joinId/request-track", requestTrack)
	router.PUT("/id/:joinId/fallback-playlist", changeFallbackPlaylist)
	router.DELETE("/id/:joinId/fallback-playlist", removeFallbackPlaylist)
}
