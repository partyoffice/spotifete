package controller

import (
	authentication "github.com/47-11/spotifete/authentication/api"
	listeningSession "github.com/47-11/spotifete/listeningSession/api"
	"github.com/gin-gonic/gin"
	"net/http"
)

type ApiV2Controller struct{ Controller }

func (controller ApiV2Controller) SetupWithBaseRouter(baseRouter *gin.Engine) {
	router := baseRouter.Group("/api/v2")

	router.GET("/", controller.Index)
	router.GET("/auth/session/new", authentication.NewSession)
	router.GET("/auth/session/id/:sessionId/is-authenticated", authentication.IsSessionAuthenticated)
	router.PATCH("/auth/session/id/:sessionId/invalidate", authentication.InvalidateSession)
	router.Any("/auth/success", authentication.CallbackSuccess)
	router.POST("/session/new", listeningSession.NewSession)
	router.GET("/session/id/:joinId", listeningSession.GetSession)
	router.DELETE("/session/id/:joinId", listeningSession.CloseSession)
	router.GET("/session/id/:joinId/search/track")
	router.GET("/session/id/:joinId/search/playlist")
	router.POST("/session/id/:joinId/request-track")
	router.GET("/session/id/:joinId/queue-last-updated")
	router.GET("/session/id/:joinId/qrcode")
	router.GET("/user/me")
	router.GET("/user/id/:userId")
}

func (ApiV2Controller) Index(c *gin.Context) {
	c.String(http.StatusOK, "Spotifete API v2")
}
