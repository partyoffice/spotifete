package webapp

import (
	"github.com/47-11/spotifete/authentication"
	"github.com/47-11/spotifete/database/model"
	. "github.com/47-11/spotifete/shared"
	"github.com/47-11/spotifete/users"
	"github.com/gin-gonic/gin"
	"net/http"
)

func SetupAuthenticationRouter(baseRouter *gin.Engine) {
	group := baseRouter.Group("/auth")

	group.GET("/callback", callback)
}

func callback(c *gin.Context) {
	loginSession, spotifeteError := getValidLoginSessionFromContext(c)
	if spotifeteError != nil {
		c.String(spotifeteError.HttpStatus, spotifeteError.MessageForUser)
		return
	}

	token, spotifeteError := authentication.GetTokenFromCallback(c)
	if spotifeteError != nil {
		c.String(spotifeteError.HttpStatus, spotifeteError.MessageForUser)
		return
	}

	_, spotifeteError = users.CreateAuthenticatedUser(token, loginSession)
	if spotifeteError != nil {
		c.String(spotifeteError.HttpStatus, spotifeteError.MessageForUser)
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
	sessionId := c.Query("state")

	session := authentication.GetSession(sessionId)
	if session == nil {
		return model.LoginSession{}, NewUserError("Unknown state.")
	}

	if !session.IsValid() {
		return model.LoginSession{}, NewUserError("Invalid state.")
	}

	if session.IsAuthenticated() {
		return model.LoginSession{}, NewUserError("State has already been used.")
	}

	return *session, nil
}
