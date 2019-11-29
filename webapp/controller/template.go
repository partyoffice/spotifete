package controller

import (
	"github.com/47-11/spotifete/service"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

type TemplateController struct{}

func (TemplateController) Index(c *gin.Context) {
	loginSession := service.LoginSessionService().GetSessionFromCookie(c)
	if loginSession == nil || loginSession.UserId == nil {
		c.HTML(http.StatusOK, "index.html", gin.H{
			"time":               time.Now(),
			"activeSessionCount": service.ListeningSessionService().GetActiveSessionCount(),
			"totalSessionCount":  service.ListeningSessionService().GetTotalSessionCount(),
			"user":               nil,
			"userSessions":       nil,
		})
		return
	}

	user := service.UserService().GetUserById(*loginSession.UserId)
	if user == nil {
		c.HTML(http.StatusOK, "index.html", gin.H{
			"time":               time.Now(),
			"activeSessionCount": service.ListeningSessionService().GetActiveSessionCount(),
			"totalSessionCount":  service.ListeningSessionService().GetTotalSessionCount(),
			"user":               nil,
			"userSessions":       nil,
		})
		return
	}

	c.HTML(http.StatusOK, "index.html", gin.H{
		"time":               time.Now(),
		"activeSessionCount": service.ListeningSessionService().GetActiveSessionCount(),
		"totalSessionCount":  service.ListeningSessionService().GetTotalSessionCount(),
		"user":               user,
		"userSessions":       service.ListeningSessionService().GetActiveSessionsByOwnerId(*loginSession.UserId),
	})
}

func (TemplateController) JoinSession(c *gin.Context) {
	joinId := c.Query("joinId")
	listeningSession := service.ListeningSessionService().GetSessionByJoinId(joinId)

	if listeningSession == nil {
		c.String(http.StatusNotFound, "session not found")
		return
	}

	c.HTML(http.StatusOK, "joinSession.html", gin.H{
		"session": listeningSession,
	})
}
