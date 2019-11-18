package controller

import (
	"github.com/47-11/spotifete/service"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

type TemplateController struct {
	sessionService     service.SessionService
	spotifyUserService service.SpotifyUserService
	spotifyService     service.SpotifyService
}

func (controller TemplateController) Index(c *gin.Context) {
	session := sessions.Default(c)
	client, err := controller.spotifyService.GetSpotifyClientUserFromSession(session)
	if err != nil || client == nil {
		c.HTML(http.StatusOK, "index.html", gin.H{
			"time":               time.Now(),
			"activeSessionCount": controller.sessionService.GetActiveSessionCount(),
			"totalSessionCount":  controller.sessionService.GetTotalSessionCount(),
			"spotifyUser":        nil,
			"userSessions":       nil,
			"authUrl":            controller.spotifyService.GetAuthUrl(),
		})

		return
	}

	// We have a client. That means we are authorized to access spotify
	spotifyUser, err := client.CurrentUser()
	if err != nil {
		c.String(http.StatusInternalServerError, "Could not get current user: "+err.Error())
	}

	internalUser, err := controller.spotifyUserService.GetOrCreateUserForSpotifyPrivateUser(spotifyUser)
	if err != nil {
		c.String(http.StatusInternalServerError, "Could not create or get internalUser: "+err.Error())
		return
	}

	userSessions := controller.sessionService.GetActiveSessionsByOwnerId(internalUser.ID)

	c.HTML(http.StatusOK, "index.html", gin.H{
		"time":               time.Now(),
		"activeSessionCount": controller.sessionService.GetActiveSessionCount(),
		"totalSessionCount":  controller.sessionService.GetTotalSessionCount(),
		"spotifyUser":        internalUser,
		"userSessions":       userSessions,
		"authUrl":            nil,
	})
}
