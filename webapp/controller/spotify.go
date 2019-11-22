package controller

import (
	"github.com/47-11/spotifete/service"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
)

type SpotifyController struct {
	spotifyService service.SpotifyService
	userService    service.UserService
}

func (controller SpotifyController) Login(c *gin.Context) {
	authUrl, _ := controller.spotifyService.NewAuthUrl()
	c.Redirect(http.StatusTemporaryRedirect, authUrl)
}

func (controller SpotifyController) Callback(c *gin.Context) {
	// Set user and token in session and redirect back to index
	spotifyService := controller.spotifyService

	state := c.Request.FormValue("state")
	err := controller.spotifyService.InvalidateState(state)
	if err != nil {
		c.String(http.StatusUnauthorized, err.Error())
		log.Println(err.Error())
		return
	}

	token, err := spotifyService.GetAuthenticator().Token(state, c.Request)
	if err != nil {
		c.String(http.StatusUnauthorized, "Could not get token: "+err.Error())
		log.Println(err.Error())
		return
	}

	// Get the spotify user for the received callback
	client := controller.spotifyService.GetAuthenticator().NewClient(token)
	spotifyUser, err := client.CurrentUser()
	if err != nil {
		c.String(http.StatusInternalServerError, "Could not get current spotify user: "+err.Error())
		log.Println(err.Error())
		return
	}

	// Get or create the database entry for the current user
	user := controller.userService.GetOrCreateUser(spotifyUser)
	controller.userService.SetToken(user, token)

	// Associate user with current session
	service.LoginSessionService().GetOrCreateSessionId(c, &user.ID)

	c.Redirect(http.StatusTemporaryRedirect, "/")
}

func (controller SpotifyController) Logout(c *gin.Context) {
	service.LoginSessionService().DeleteSessionId(c)

	c.Redirect(http.StatusTemporaryRedirect, "/")
}
