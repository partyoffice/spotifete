package controller

import (
	"github.com/47-11/spotifete/service"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

type TemplateController struct {
	listeningSessionService service.ListeningSessionService
	userService             service.UserService
	spotifyService          service.SpotifyService
}

func (controller TemplateController) Index(c *gin.Context) {
	loginSession := service.LoginSessionService().GetSessionFromCookie(c)
	if loginSession == nil || loginSession.UserId == nil {
		c.HTML(http.StatusOK, "index.html", gin.H{
			"time":               time.Now(),
			"activeSessionCount": controller.listeningSessionService.GetActiveSessionCount(),
			"totalSessionCount":  controller.listeningSessionService.GetTotalSessionCount(),
			"user":               nil,
			"userSessions":       nil,
		})
		return
	}

	user, err := controller.userService.GetUserById(*loginSession.UserId)
	if err != nil {
		c.HTML(http.StatusOK, "index.html", gin.H{
			"time":               time.Now(),
			"activeSessionCount": controller.listeningSessionService.GetActiveSessionCount(),
			"totalSessionCount":  controller.listeningSessionService.GetTotalSessionCount(),
			"user":               nil,
			"userSessions":       nil,
		})
		return
	}

	c.HTML(http.StatusOK, "index.html", gin.H{
		"time":               time.Now(),
		"activeSessionCount": controller.listeningSessionService.GetActiveSessionCount(),
		"totalSessionCount":  controller.listeningSessionService.GetTotalSessionCount(),
		"user":               user,
		"userSessions":       controller.listeningSessionService.GetActiveSessionsByOwnerId(*loginSession.UserId),
	})
}
