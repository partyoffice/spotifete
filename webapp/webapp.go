package webapp

import (
	. "github.com/47-11/spotifete/webapp/controller"
	"github.com/gin-gonic/gin"
)

func Start(activeProfile string) {
	gin.SetMode(activeProfile)
	ginEngine := gin.Default()

	registerRoutes(ginEngine)

	ginEngine.Run(":8410")
}

func registerRoutes(baseRouter *gin.Engine) {
	// Templates
	baseRouter.LoadHTMLGlob("resources/templates/*.html")
	templateController := new(TemplateController)
	baseRouter.GET("/", templateController.Index)

	// API
	apiRouter := baseRouter.Group("/api/v1")
	apiController := new(ApiController)

	apiRouter.GET("/", apiController.Index)
	apiRouter.GET("/sessions", apiController.GetActiveSessions)
	apiRouter.GET("/sessions/:sessionId", apiController.GetSession)
	apiRouter.GET("/users/:userId", apiController.GetUser)
}
