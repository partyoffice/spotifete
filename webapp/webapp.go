package webapp

import (
	. "github.com/47-11/spotifete/webapp/controller"
	sentrygin "github.com/getsentry/sentry-go/gin"
	"github.com/gin-gonic/gin"
	"log"
)

func Start(activeProfile string) {
	gin.SetMode(activeProfile)
	baseRouter := gin.Default()

	baseRouter.Use(sentrygin.New(sentrygin.Options{
		Repanic: true,
	}))

	setupStaticController(baseRouter)
	setupApiController(baseRouter)
	setupTemplateController(baseRouter)
	setupSpotifyController(baseRouter)

	err := baseRouter.Run(":8410")

	if err != nil {
		log.Fatalln(err.Error())
	}
}

func setupStaticController(baseRouter *gin.Engine) {
	baseRouter.Static("/static/", "./resources/static/")
}

func setupApiController(baseRouter *gin.Engine) {
	apiRouter := baseRouter.Group("/api/v1")
	apiController := new(ApiController)

	apiRouter.GET("/", apiController.Index)
	apiRouter.GET("/spotify/auth/new", apiController.GetAuthUrl)
	apiRouter.GET("/spotify/auth/authenticated", apiController.DidAuthSucceed)
	apiRouter.PATCH("/spotify/auth/invalidate", apiController.InvalidateSessionId)
	apiRouter.GET("/spotify/search/track", apiController.SearchSpotifyTrack)
	apiRouter.GET("/sessions/:joinId", apiController.GetSession)
	apiRouter.POST("/sessions/:joinId/request", apiController.RequestSong)
	apiRouter.POST("/sessions", apiController.CreateListeningSession)
	apiRouter.GET("/users/:userId", apiController.GetUser)
}

func setupTemplateController(baseRouter *gin.Engine) {
	baseRouter.LoadHTMLGlob("resources/templates/*.html")
	templateController := new(TemplateController)
	baseRouter.GET("/", templateController.Index)
	baseRouter.POST("/", templateController.Index)
	baseRouter.GET("/session/view/:joinId", templateController.ViewSession)
	baseRouter.POST("/session/view/:joinId", templateController.ViewSession)
	baseRouter.GET("/session/new", templateController.NewListeningSession)
	baseRouter.POST("/session/new", templateController.NewListeningSessionSubmit)
	baseRouter.POST("session/close", templateController.CloseListeningSession)
}

func setupSpotifyController(baseRouter *gin.Engine) {
	spotifyRouter := baseRouter.Group("/spotify")
	spotifyController := new(SpotifyController)

	spotifyRouter.GET("/login", spotifyController.Login)
	spotifyRouter.GET("/callback", spotifyController.Callback)
	spotifyRouter.GET("/logout", spotifyController.Logout)
}
