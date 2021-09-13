package authentication

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/partyoffice/spotifete/authentication"
	. "github.com/partyoffice/spotifete/webapp/apiv2/shared"
)

func newSession(c *gin.Context) {
	callbackRedirectUrl := c.DefaultQuery("redirectTo", "/api/v2/auth/success")

	session, spotifyAuthenticationUrl, spotifeteError := authentication.NewSession(callbackRedirectUrl)
	if spotifeteError == nil {
		c.JSON(http.StatusOK, NewSessionResponse{
			SpotifyAuthenticationUrl: spotifyAuthenticationUrl,
			SpotifeteSessionId:       session.SessionId,
		})
	} else {
		SetJsonError(*spotifeteError, c)
	}
}

func isSessionAuthenticated(c *gin.Context) {
	sessionId := c.Param("sessionId")
	session := authentication.GetSession(sessionId)
	if session == nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Message: "Unknown session id."})
		return
	}

	authenticated := session.IsAuthenticated()
	response := IsSessionAuthenticatedResponse{Authenticated: authenticated}
	if authenticated {
		c.JSON(http.StatusOK, response)
	} else {
		c.JSON(http.StatusUnauthorized, response)
	}
}

func invalidateSession(c *gin.Context) {
	sessionId := c.Param("sessionId")
	authentication.InvalidateSession(sessionId)

	c.Status(http.StatusNoContent)
}

func callbackSuccess(c *gin.Context) {
	// TODO: Do something nicer here
	c.String(http.StatusOK, "Authentication successful! You can close this window now.")
}
