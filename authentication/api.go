package authentication

import (
	"github.com/47-11/spotifete/authentication/model/api"
	"github.com/gin-gonic/gin"
	"net/http"
)

func ApiNewSession(c *gin.Context) {
	session, spotifyAuthenticationUrl := NewSession("/api/v2/auth/success")

	c.JSON(http.StatusOK, api.NewSessionResponse{
		SpotifyAuthenticationUrl: spotifyAuthenticationUrl,
		SpotifeteSessionId:       session.SessionId,
	})
}

func ApiIsSessionAuthenticated(c *gin.Context) {
	sessionId := c.Param("sessionId")
	sessionAuthenticated, spotifeteError := isSessionAuthenticatedBySessionId(sessionId)
	if spotifeteError != nil {
		spotifeteError.SetJsonResponse(c)
		return
	}

	response := api.IsSessionAuthenticatedResponse{Authenticated: sessionAuthenticated}
	if sessionAuthenticated {
		c.JSON(http.StatusOK, response)
	} else {
		c.JSON(http.StatusUnauthorized, response)
	}
}

func ApiInvalidateSession(c *gin.Context) {
	sessionId := c.Param("sessionId")
	InvalidateSession(sessionId)

	c.Status(http.StatusNoContent)
}

func ApiCallbackSuccess(c *gin.Context) {
	// TODO: Do something nicer here
	c.String(http.StatusOK, "Authentication successful! You can close this window now.")
}
