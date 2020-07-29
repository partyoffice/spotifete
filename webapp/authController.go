package webapp

import (
	"github.com/47-11/spotifete/service"
	"github.com/getsentry/sentry-go"
	"github.com/gin-gonic/gin"
	"github.com/google/logger"
	"net/http"
)

type AuthController struct{}

func (c AuthController) SetupRoutes(baseRouter *gin.Engine) {
	router := baseRouter.Group("/spotify")

	router.GET("/login", c.SpotifyLogin)
	router.GET("/logout", c.Logout)
	router.GET("/callback", c.SpotifyCallback)
	router.GET("/apicallback", c.SpotifyApiCallback)
}

func (AuthController) SpotifyLogin(c *gin.Context) {
	redirectTo := c.Query("redirectTo")
	if len(redirectTo) == 0 {
		redirectTo = "/"
	}

	authUrl, _ := service.SpotifyService().NewAuthUrl(redirectTo)
	c.Redirect(http.StatusTemporaryRedirect, authUrl)
}

func (AuthController) SpotifyCallback(c *gin.Context) {
	// Set user and token in session and redirect back to index
	state := c.Request.FormValue("state")

	// Check that this state exists and was not used in a callback before
	session := service.LoginSessionService().GetSessionBySessionId(state, true)
	if session == nil {
		c.String(http.StatusUnauthorized, "invalid state.")
		return
	}

	if session.UserId != nil {
		c.String(http.StatusUnauthorized, "State has already been used.")
		return
	}

	// Fetch the token
	token, err := service.SpotifyService().Authenticator.Token(state, c.Request)
	if err != nil {
		logger.Error(err)
		sentry.CaptureException(err)
		c.String(http.StatusInternalServerError, "Could not get token: "+err.Error())
		return
	}

	// Get the spotify user for the token
	client := service.SpotifyService().Authenticator.NewClient(token)
	spotifyUser, err := client.CurrentUser()
	if err != nil {
		logger.Error(err)
		sentry.CaptureException(err)
		c.String(http.StatusInternalServerError, "Could not get current spotify user: "+err.Error())
		return
	}

	// Cache the created client
	service.SpotifyService().Clients[spotifyUser.ID] = &client

	// Get or create the database entry for the current user
	user := service.UserService().GetOrCreateUser(spotifyUser)
	service.UserService().SetToken(user, *token)

	// Associate user with current session
	service.LoginSessionService().SetUserForSession(*session, user)

	// Set or update session cookie
	service.LoginSessionService().SetSessionCookie(c, session.SessionId)

	redirectTo := session.CallbackRedirect
	if redirectTo[0:1] != "/" {
		redirectTo = "/" + redirectTo
	}
	c.Redirect(http.StatusSeeOther, redirectTo)
}

func (AuthController) SpotifyApiCallback(c *gin.Context) {
	// TODO: Do something nicer here
	c.String(http.StatusOK, "Logged in successfully! You can close this window.")
}

func (AuthController) Logout(c *gin.Context) {
	_ = service.LoginSessionService().InvalidateSession(c)

	redirectTo := c.Query("redirectTo")
	if len(redirectTo) == 0 || redirectTo[0:1] != "/" {
		redirectTo = "/" + redirectTo
	}

	c.Redirect(http.StatusTemporaryRedirect, redirectTo)
}
