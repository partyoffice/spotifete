package api

import (
	"github.com/47-11/spotifete/authentication"
	"github.com/gin-gonic/gin"
	"net/http"
)

func newSession(c *gin.Context) {
	session, spotifyAuthenticationUrl := authentication.NewSession("/api/v2/auth/success")

	c.JSON(http.StatusOK, NewSessionResponse{
		SpotifyAuthenticationUrl: spotifyAuthenticationUrl,
		SpotifeteSessionId:       session.SessionId,
	})
}

func isSessionAuthenticated(c *gin.Context) {
	sessionId := c.Param("sessionId")
	sessionAuthenticated, spotifeteError := authentication.IsSessionAuthenticatedBySessionId(sessionId)
	if spotifeteError != nil {
		spotifeteError.SetJsonResponse(c)
		return
	}

	response := IsSessionAuthenticatedResponse{Authenticated: sessionAuthenticated}
	if sessionAuthenticated {
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
