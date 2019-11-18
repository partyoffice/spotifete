package webapp

import (
	"crypto/rand"
	. "github.com/47-11/spotifete/webapp/controller"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
)

func Start(activeProfile string) {
	gin.SetMode(activeProfile)
	baseRouter := gin.Default()

	setupSessions(baseRouter)

	setupApiController(baseRouter)
	setupTemplateController(baseRouter)
	setupSpotifyController(baseRouter)

	baseRouter.Run(":8410")
}

func setupSessions(baseRouter *gin.Engine) {
	randomBytes := make([]byte, 16)
	_, err := rand.Read(randomBytes)
	if err != nil {
		panic("Could not create random randomBytes for cookie store: " + err.Error())
	}

	baseRouter.Use(sessions.Sessions("spotifetesession", cookie.NewStore(randomBytes)))
}

func setupApiController(baseRouter *gin.Engine) {
	apiRouter := baseRouter.Group("/api/v1")
	apiController := new(ApiController)

	apiRouter.GET("/", apiController.Index)
	apiRouter.GET("/sessions", apiController.GetActiveSessions)
	apiRouter.GET("/sessions/:sessionId", apiController.GetSession)
	apiRouter.GET("/users/:userId", apiController.GetUser)
}

func setupTemplateController(baseRouter *gin.Engine) {
	baseRouter.LoadHTMLGlob("resources/templates/*.html")
	templateController := new(TemplateController)
	baseRouter.GET("/", templateController.Index)
}

func setupSpotifyController(baseRouter *gin.Engine) {
	spotifyRouter := baseRouter.Group("/spotify")
	spotifyController := new(SpotifyController)

	spotifyRouter.GET("/login", spotifyController.Login)
	spotifyRouter.GET("/callback", spotifyController.Callback)
}
