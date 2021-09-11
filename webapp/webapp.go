package webapp

import (
	"fmt"

	"github.com/getsentry/sentry-go"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/google/logger"
	"github.com/partyoffice/spotifete/config"
	"github.com/partyoffice/spotifete/webapp/apiv2"
)

type SpotifeteWebapp struct {
	router *gin.Engine
}

func (w SpotifeteWebapp) Setup() SpotifeteWebapp {
	w = w.createAndConfigureRouter()
	w = w.setupCors()
	w.setupRoutes()

	return w
}

func (w SpotifeteWebapp) createAndConfigureRouter() SpotifeteWebapp {
	w.setGinMode()
	w.router = gin.Default()

	return w
}

func (SpotifeteWebapp) setGinMode() {
	c := config.Get()
	if c.SpotifeteConfiguration.ReleaseMode {
		logger.Infof("Running in release mode on port %d", c.SpotifeteConfiguration.Port)
		gin.SetMode(gin.ReleaseMode)
	} else {
		logger.Infof("Running in debug mode on port %d", c.SpotifeteConfiguration.Port)
		gin.SetMode(gin.DebugMode)
	}
}

func (w SpotifeteWebapp) setupCors() SpotifeteWebapp {
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowAllOrigins = true
	w.router.Use(cors.New(corsConfig))

	return w
}

func (w SpotifeteWebapp) setupRoutes() {
	SetupStaticRouter(w.router)
	SetupAuthenticationRouter(w.router)
	apiv2.SetupApiRoutes(w.router)

	TemplateController{}.SetupWithBaseRouter(w.router)
}

func (w SpotifeteWebapp) Run() {
	err := w.router.Run(fmt.Sprintf(":%d", config.Get().SpotifeteConfiguration.Port))

	if err != nil {
		sentry.CaptureException(err)
		logger.Fatal(err)
	}
}
