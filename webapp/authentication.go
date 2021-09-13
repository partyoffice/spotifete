package webapp

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/partyoffice/spotifete/authentication"
	"github.com/partyoffice/spotifete/database/model"
	. "github.com/partyoffice/spotifete/shared"
	"github.com/partyoffice/spotifete/users"
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
