package api

import (
	"github.com/gin-gonic/gin"
)

func SetupRouter(baseRouterGroup *gin.RouterGroup) {
	router := baseRouterGroup.Group("/session")

	router.POST("/new", newSession)
	router.GET("/id/:joinId", getSession)
	router.DELETE("/id/:joinId", closeSession)
	router.GET("/id/:joinId/search/track")
	router.GET("/id/:joinId/search/playlist")
	router.POST("/id/:joinId/request-track")
	router.GET("/id/:joinId/queue-last-updated")
	router.GET("/id/:joinId/qrcode")
}
