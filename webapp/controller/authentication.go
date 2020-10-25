package controller

import (
	"github.com/47-11/spotifete/authentication"
	"github.com/47-11/spotifete/database/model"
	. "github.com/47-11/spotifete/error"
	"github.com/47-11/spotifete/users"
	"github.com/gin-gonic/gin"
	"net/http"
)

type OAuth2AuthenticationController interface {
	Controller
	Callback(*gin.Context)
}

type SpotifyAuthenticationController struct{ OAuth2AuthenticationController }

func (c SpotifyAuthenticationController) SetupWithBaseRouter(baseRouter *gin.Engine) {
	group := baseRouter.Group("/auth")

	group.GET("/callback", c.Callback)
}

func (SpotifyAuthenticationController) Callback(c *gin.Context) {
	loginSession, spotifeteError := getValidLoginSessionFromContext(c)
	if spotifeteError != nil {
		spotifeteError.SetStringResponse(c)
		return
	}

	token, spotifeteError := authentication.GetTokenFromCallback(c)
	if spotifeteError != nil {
		spotifeteError.SetStringResponse(c)
		return
	}

	_, spotifeteError = users.CreateAuthenticatedUser(token, loginSession)
	if spotifeteError != nil {
		spotifeteError.SetStringResponse(c)
		return
	}

	// Set or update session cookie
	authentication.SetCookie(c, loginSession.SessionId)

	redirectTo := loginSession.CallbackRedirect
	if redirectTo[0:1] != "/" {
		redirectTo = "/" + redirectTo
	}

	c.Redirect(http.StatusSeeOther, redirectTo)
}

func getValidLoginSessionFromContext(c *gin.Context) (model.LoginSession, *SpotifeteError) {
	state := c.Query("state")

	session := authentication.GetValidSession(state)
	if session == nil {
		return model.LoginSession{}, NewUserError("Unknown state.")
	}

	if session.UserId != nil {
		return model.LoginSession{}, NewUserError("State has already been used.")
	}

	return *session, nil
}
