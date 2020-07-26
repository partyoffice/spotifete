package webapp

import (
	"fmt"
	"github.com/47-11/spotifete/config"
	. "github.com/47-11/spotifete/webapp/controller"
	"github.com/getsentry/sentry-go"
	sentrygin "github.com/getsentry/sentry-go/gin"
	"github.com/gin-gonic/gin"
	"github.com/google/logger"
	"io"
	"os"
)

func Initialize() {
	c := config.Get()
	if c.SpotifeteConfiguration.ReleaseMode {
		logger.Infof("Running in release mode on port %d", c.SpotifeteConfiguration.Port)
		gin.SetMode(gin.ReleaseMode)
	} else {
		logger.Infof("Running in debug mode on port %d", c.SpotifeteConfiguration.Port)
		gin.SetMode(gin.DebugMode)
	}

	baseRouter := gin.Default()

	// Setup logging for gin
	logFile, err := os.OpenFile("gin.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0660)
	if err != nil {
		logger.Fatalf("Failed to open gin log file: %v", err)
	}
	defer logFile.Close()
	gin.DefaultWriter = io.MultiWriter(logFile, os.Stdout)

	baseRouter.Use(sentrygin.New(sentrygin.Options{
		Repanic: true,
	}))

	// Setup routers
	setupStaticController(baseRouter)
	setupApiController(baseRouter)
	setupTemplateController(baseRouter)
	setupSpotifyController(baseRouter)

	err = baseRouter.Run(fmt.Sprintf(":%d", c.SpotifeteConfiguration.Port))

	if err != nil {
		sentry.CaptureException(err)
		logger.Fatal(err)
	}
}

func setupStaticController(baseRouter *gin.Engine) {
	baseRouter.Static("/static/", "./resources/static/")
	baseRouter.StaticFile("favicon.ico", "./resources/favicon.ico")
}

func setupApiController(baseRouter *gin.Engine) {
	apiRouter := baseRouter.Group("/api/v1")
	apiController := new(ApiController)

	apiRouter.GET("/", apiController.Index)
	apiRouter.GET("/spotify/auth/new", apiController.GetAuthUrl)
	apiRouter.GET("/spotify/auth/authenticated", apiController.DidAuthSucceed)
	apiRouter.PATCH("/spotify/auth/invalidate", apiController.InvalidateSessionId)
	apiRouter.GET("/spotify/search/track", apiController.SearchSpotifyTrack)
	apiRouter.GET("/spotify/search/playlist", apiController.SearchSpotifyPlaylist)
	apiRouter.GET("/sessions/:joinId", apiController.GetSession)
	apiRouter.DELETE("sessions/:joinId", apiController.CloseListeningSession)
	apiRouter.POST("/sessions/:joinId/request", apiController.RequestSong)
	apiRouter.GET("/sessions/:joinId/queuelastupdated", apiController.QueueLastUpdated)
	apiRouter.GET("/sessions/:joinId/qrcode", apiController.CreateQrCodeForListeningSession)
	apiRouter.POST("/sessions", apiController.CreateListeningSession)
	apiRouter.GET("/users/:userId", apiController.GetUser)
}

func setupTemplateController(baseRouter *gin.Engine) {
	baseRouter.LoadHTMLGlob("resources/templates/*.html")
	templateController := new(TemplateController)
	baseRouter.GET("/", templateController.Index)
	baseRouter.GET("/session/new", templateController.NewListeningSession)
	baseRouter.POST("/session/new", templateController.NewListeningSessionSubmit)
	baseRouter.GET("/session/view/:joinId", templateController.ViewSession)
	baseRouter.POST("/session/view/:joinId/request", templateController.RequestTrack)
	baseRouter.POST("/session/view/:joinId/fallback", templateController.ChangeFallbackPlaylist)
	baseRouter.POST("/session/close", templateController.CloseListeningSession)
	baseRouter.GET("/app", templateController.GetApp)
	baseRouter.GET("/app/android", templateController.GetAppAndroid)
	baseRouter.GET("/app/ios", templateController.GetAppIOS)
	baseRouter.GET("/apicallback")
}

func setupSpotifyController(baseRouter *gin.Engine) {
	spotifyRouter := baseRouter.Group("/spotify")
	spotifyController := new(SpotifyController)

	spotifyRouter.GET("/login", spotifyController.Login)
	spotifyRouter.GET("/callback", spotifyController.Callback)
	spotifyRouter.GET("/logout", spotifyController.Logout)
	spotifyRouter.GET("/apicallback", spotifyController.ApiCallback)
}
