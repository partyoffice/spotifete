package controller

import (
	"github.com/47-11/spotifete/authentication"
	"github.com/47-11/spotifete/database/model"
	. "github.com/47-11/spotifete/error"
	"github.com/47-11/spotifete/service"
	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
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

func getTokenFromContext(c *gin.Context) (*oauth2.Token, *SpotifeteError) {
	state := c.Query("state")

	token, err := service.SpotifyService().Authenticator.Token(state, c.Request)
	if err != nil {
		return nil, NewError("Could not fetch access token from Spotify.", err, http.StatusUnauthorized)
	}

	return token, nil
}

func authenticateUser(token *oauth2.Token, session model.LoginSession) *SpotifeteError {
	client := service.SpotifyService().Authenticator.NewClient(token)
	spotifyUser, err := client.CurrentUser()
	if err != nil {
		return NewError("Could not get user information from Spotify.", err, http.StatusInternalServerError)
	}

	service.SpotifyService().Clients[spotifyUser.ID] = &client

	user := service.UserService().GetOrCreateUser(spotifyUser)
	service.UserService().SetToken(user, *token)

	authentication.SetUserForSession(session, user)

	return nil
}
