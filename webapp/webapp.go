package webapp

import (
	"fmt"
	"github.com/47-11/spotifete/config"
	"github.com/47-11/spotifete/webapp/apiv1"
	"github.com/47-11/spotifete/webapp/apiv2"
	"github.com/getsentry/sentry-go"
	sentrygin "github.com/getsentry/sentry-go/gin"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/google/logger"
	"io"
	"os"
)

type SpotifeteWebapp struct {
	router  *gin.Engine
	logFile *os.File
}

func (w SpotifeteWebapp) Setup() SpotifeteWebapp {
	w = w.createAndConfigureRouter()
	w = w.setupLogging()
	w = w.setupCors()
	w.setupRoutes()

	return w
}

func (w SpotifeteWebapp) createAndConfigureRouter() SpotifeteWebapp {
	w.setGinModeDependingOnConfiguration()
	w.router = gin.Default()

	return w
}

func (SpotifeteWebapp) setGinModeDependingOnConfiguration() {
	c := config.Get()
	if c.SpotifeteConfiguration.ReleaseMode {
		logger.Infof("Running in release mode on port %d", c.SpotifeteConfiguration.Port)
		gin.SetMode(gin.ReleaseMode)
	} else {
		logger.Infof("Running in debug mode on port %d", c.SpotifeteConfiguration.Port)
		gin.SetMode(gin.DebugMode)
	}
}

func (w SpotifeteWebapp) setupLogging() SpotifeteWebapp {
	w = w.setupGinLogging()
	w.setupSentryLogging()

	return w
}

func (w SpotifeteWebapp) setupGinLogging() SpotifeteWebapp {
	logFile, err := os.OpenFile("gin.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0660)
	if err != nil {
		logger.Fatalf("Failed to open gin log file: %v", err)
	}

	w.logFile = logFile

	gin.DefaultWriter = io.MultiWriter(logFile, os.Stdout)

	return w
}

func (w SpotifeteWebapp) setupSentryLogging() {
	w.router.Use(sentrygin.New(sentrygin.Options{
		Repanic: true,
	}))
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
	apiv1.ApiV1Controller{}.SetupWithBaseRouter(w.router)
}

func (w SpotifeteWebapp) Run() {
	err := w.router.Run(fmt.Sprintf(":%d", config.Get().SpotifeteConfiguration.Port))

	if err != nil {
		sentry.CaptureException(err)
		logger.Fatal(err)
	}
}

func (w SpotifeteWebapp) Shutdown() {
	err := w.logFile.Close()
	if err != nil {
		panic("Could not close gin log file: " + err.Error())
	}
}
