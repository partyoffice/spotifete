package controller

import (
"github.com/47-11/spotifete/service"
"github.com/getsentry/sentry-go"
"github.com/gin-gonic/gin"
"github.com/google/logger"
"net/http"
)

type AuthenticationController interface {
	Controller
	Login(*gin.Context)
	Logout(*gin.Context)
}

type OAuth2AuthenticationController interface {
	AuthenticationController
	Callback(*gin.Context)
}


type SpotifyAuthenticationController struct{OAuth2AuthenticationController}

func (c SpotifyAuthenticationController) SetupWithBaseRouter(baseRouter *gin.Engine) {
	group := baseRouter.Group("/spotify")
	group.GET("/login", c.Login)
	group.GET("/logout", c.Logout)
	group.GET("/callback", c.Callback)
}

func (SpotifyAuthenticationController) Login(c *gin.Context) {
	redirectTo := c.DefaultQuery("redirectTo", "/")

	authUrl, _ := service.SpotifyService().NewAuthUrl(redirectTo)
	c.Redirect(http.StatusTemporaryRedirect, authUrl)
}

func (SpotifyAuthenticationController) Callback(c *gin.Context) {
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

func (SpotifyAuthenticationController) SpotifyApiCallback(c *gin.Context) {
	// TODO: Do something nicer here
	c.String(http.StatusOK, "Logged in successfully! You can close this window.")
}

func (SpotifyAuthenticationController) Logout(c *gin.Context) {
	_ = service.LoginSessionService().InvalidateSession(c)

	redirectTo := c.Query("redirectTo")
	if len(redirectTo) == 0 || redirectTo[0:1] != "/" {
		redirectTo = "/" + redirectTo
	}

	c.Redirect(http.StatusTemporaryRedirect, redirectTo)
}

