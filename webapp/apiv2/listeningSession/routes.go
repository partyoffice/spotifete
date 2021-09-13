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
	router.DELETE("/id/:joinId/queue", deleteRequestFromQueue)
	router.GET("/id/:joinId/queue/last-updated", queueLastUpdated)
	router.GET("/id/:joinId/qrcode", qrCode)
	router.GET("/id/:joinId/search/track", searchTrack)
	router.GET("/id/:joinId/search/playlist", searchPlaylist)
	router.POST("/id/:joinId/request-track", requestTrack)
	router.POST("/id/:joinId/new-queue-playlist", newQueuePlaylist)
	router.PUT("/id/:joinId/fallback-playlist", changeFallbackPlaylist)
	router.DELETE("/id/:joinId/fallback-playlist", removeFallbackPlaylist)
	router.PATCH("/id/:joinId/fallback-playlist/shuffle", setFallbackPlaylistShuffle)
}
