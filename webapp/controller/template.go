package controller

import (
	"github.com/47-11/spotifete/service"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

type TemplateController struct {
	sessionService service.ListeningSessionService
	userService    service.UserService
	spotifyService service.SpotifyService
}

func (controller TemplateController) Index(c *gin.Context) {
	_, userId := service.LoginSessionService().GetOrCreateSessionId(c, nil)
	if userId == nil {
		c.HTML(http.StatusOK, "index.html", gin.H{
			"time":               time.Now(),
			"activeSessionCount": controller.sessionService.GetActiveSessionCount(),
			"totalSessionCount":  controller.sessionService.GetTotalSessionCount(),
			"user":               nil,
			"userSessions":       nil,
			"authUrl":            controller.spotifyService.GetAuthUrl(),
		})
		return
	}

	user, err := controller.userService.GetUserById(*userId)
	if err != nil {
		c.HTML(http.StatusOK, "index.html", gin.H{
			"time":               time.Now(),
			"activeSessionCount": controller.sessionService.GetActiveSessionCount(),
			"totalSessionCount":  controller.sessionService.GetTotalSessionCount(),
			"user":               nil,
			"userSessions":       nil,
			"authUrl":            controller.spotifyService.GetAuthUrl(),
		})
		return
	}

	c.HTML(http.StatusOK, "index.html", gin.H{
		"time":               time.Now(),
		"activeSessionCount": controller.sessionService.GetActiveSessionCount(),
		"totalSessionCount":  controller.sessionService.GetTotalSessionCount(),
		"user":               user,
		"userSessions":       controller.sessionService.GetActiveSessionsByOwnerId(*userId),
		"authUrl":            nil,
	})
}
