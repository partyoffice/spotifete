package webapp

import (
	"fmt"
	"github.com/47-11/spotifete/config"
	"github.com/getsentry/sentry-go"
	sentrygin "github.com/getsentry/sentry-go/gin"
	"github.com/gin-gonic/gin"
	"github.com/google/logger"
	"io"
	"os"
)

type Controller interface {
	SetupRoutes(BaseRouter *gin.Engine)
}

type Webapp interface {
	Setup()
	Run()
	Shutdown()
}

type SpotifeteWebapp struct {
	router  *gin.Engine
	logFile *os.File
}

func (w SpotifeteWebapp) Setup() {
	w.createAndConfigureRouter()
	w.setupLogging()
	w.setupRoutes()
}

func (w SpotifeteWebapp) createAndConfigureRouter() {
	w.setGinModeDependingOnConfiguration()
	w.router = gin.Default()
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

func (w SpotifeteWebapp) setupLogging() {
	w.setupGinLogging()
	w.setupSentryLogging()
}

func (w SpotifeteWebapp) setupGinLogging() {
	logFile, err := os.OpenFile("gin.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0660)
	if err != nil {
		logger.Fatalf("Failed to open gin log file: %v", err)
	}

	gin.DefaultWriter = io.MultiWriter(logFile, os.Stdout)
}

func (w SpotifeteWebapp) setupSentryLogging() {
	w.router.Use(sentrygin.New(sentrygin.Options{
		Repanic: true,
	}))
}

func (w SpotifeteWebapp) setupRoutes() {
	AuthController{}.SetupRoutes(w.router)
	StaticController{}.SetupRoutes(w.router)
	TemplateController{}.SetupRoutes(w.router)
	ApiController{}.SetupRoutes(w.router)
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
