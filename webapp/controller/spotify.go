package controller

import (
	"github.com/47-11/spotifete/service"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"time"
)

type SpotifyController struct {
	spotifyService service.SpotifyService
}

func (controller SpotifyController) Login(c *gin.Context) {
	c.Redirect(http.StatusTemporaryRedirect, controller.spotifyService.GetAuthUrl())
}

func (controller SpotifyController) Callback(c *gin.Context) {
	// Set user and token in session and redirect back to index
	spotifyService := controller.spotifyService
	state := spotifyService.GetState()

	token, err := spotifyService.GetAuthenticator().Token(state, c.Request)
	if err != nil {
		c.String(http.StatusForbidden, "Could not get token: "+err.Error())
		log.Println(err.Error())
		return
	}
	if st := c.Request.FormValue("state"); st != state {
		c.String(http.StatusUnauthorized, "State mismatch")
		log.Printf("State mismatch: %s != %s\n", st, state)
		return
	}

	// Save token on session
	session := sessions.Default(c)
	session.Set("spotifyAccesstoken", token.AccessToken)
	session.Set("spotifyRefreshToken", token.RefreshToken)
	session.Set("spotifyTokenExpiry", token.Expiry.Format(time.RFC3339))
	session.Set("spotifyTokenType", token.TokenType)
	err = session.Save()

	if err != nil {
		c.String(http.StatusInternalServerError, "Could not save session: "+err.Error())
		log.Println(err)
		return
	}

	// TODO: Add something like authSuccess / error as Redirect parameter and display success
	c.Redirect(http.StatusTemporaryRedirect, "/")
}
