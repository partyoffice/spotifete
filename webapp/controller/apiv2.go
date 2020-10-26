package controller

import (
	authentication "github.com/47-11/spotifete/authentication/api"
	listeningSession "github.com/47-11/spotifete/listeningSession/api"
	users "github.com/47-11/spotifete/users/api"
	"github.com/gin-gonic/gin"
	"net/http"
)

type ApiV2Controller struct{ Controller }

func (controller ApiV2Controller) SetupWithBaseRouter(baseRouter *gin.Engine) {
	router := baseRouter.Group("/api/v2")

	router.GET("/", controller.Index)
	authentication.SetupRouter(router)
	listeningSession.SetupRouter(router)
	users.SetupRouter(router)
}

func (ApiV2Controller) Index(c *gin.Context) {
	c.String(http.StatusOK, "Spotifete API v2")
}
