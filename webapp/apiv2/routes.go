package apiv2

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/partyoffice/spotifete/webapp/apiv2/authentication"
	"github.com/partyoffice/spotifete/webapp/apiv2/listeningSession"
	"github.com/partyoffice/spotifete/webapp/apiv2/user"
)

func SetupApiRoutes(baseRouter *gin.Engine) {
	router := baseRouter.Group("/api/v2")

	router.GET("/", index)

	authentication.SetupRoutes(router)
	listeningSession.SetupRoutes(router)
	user.SetupRoutes(router)
}

func index(c *gin.Context) {
	c.String(http.StatusOK, "Spotifete API v2. This api is still WIP and not completely implemented yet.")
}
