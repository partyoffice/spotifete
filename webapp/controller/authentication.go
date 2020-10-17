package controller

import (
	. "github.com/47-11/spotifete/error"
	"github.com/47-11/spotifete/model/database"
	"github.com/47-11/spotifete/service"
	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
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

type SpotifyAuthenticationController struct{ OAuth2AuthenticationController }

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
	loginSession, spotifeteError := getValidLoginSessionFromContext(c)
	if spotifeteError != nil {
		spotifeteError.SetStringResponse(c)
		return
	}

	token, spotifeteError := getTokenFromContext(c)
	if spotifeteError != nil {
		spotifeteError.SetStringResponse(c)
		return
	}

	spotifeteError = authenticateUser(token, loginSession)
	if spotifeteError != nil {
		spotifeteError.SetStringResponse(c)
		return
	}

	// Set or update session cookie
	service.LoginSessionService().SetSessionCookie(c, loginSession.SessionId)

	redirectTo := loginSession.CallbackRedirect
	if redirectTo[0:1] != "/" {
		redirectTo = "/" + redirectTo
	}

	c.Redirect(http.StatusSeeOther, redirectTo)
}

func getValidLoginSessionFromContext(c *gin.Context) (database.LoginSession, *SpotifeteError) {
	state := c.Query("state")

	session := service.LoginSessionService().GetSessionBySessionId(state, true)
	if session == nil {
		return database.LoginSession{}, NewUserError("Unknown state.")
	}

	if session.UserId != nil {
		return database.LoginSession{}, NewUserError("State has already been used.")
	}

	return *session, nil
}

func getTokenFromContext(c *gin.Context) (*oauth2.Token, *SpotifeteError) {
	state := c.Query("state")

	token, err := service.SpotifyService().Authenticator.Token(state, c.Request)
	if err != nil {
		return nil, NewError("Could not fetch access token from Spotify.", err, http.StatusUnauthorized)
	}

	return token, nil
}

func authenticateUser(token *oauth2.Token, session database.LoginSession) *SpotifeteError {
	client := service.SpotifyService().Authenticator.NewClient(token)
	spotifyUser, err := client.CurrentUser()
	if err != nil {
		return NewError("Could not get user information from Spotify.", err, http.StatusInternalServerError)
	}

	service.SpotifyService().Clients[spotifyUser.ID] = &client

	user := service.UserService().GetOrCreateUser(spotifyUser)
	service.UserService().SetToken(user, *token)

	service.LoginSessionService().SetUserForSession(session, user)

	return nil
}

func (SpotifyAuthenticationController) SpotifyApiCallback(c *gin.Context) {
	// TODO: Do something nicer here
	c.String(http.StatusOK, "Logged in successfully! You can close this window now :)")
}

func (SpotifyAuthenticationController) Logout(c *gin.Context) {
	service.LoginSessionService().InvalidateSession(c)

	redirectTo := c.Query("redirectTo")
	if len(redirectTo) == 0 || redirectTo[0:1] != "/" {
		redirectTo = "/" + redirectTo
	}

	c.Redirect(http.StatusTemporaryRedirect, redirectTo)
}
