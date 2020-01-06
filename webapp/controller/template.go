package controller

import (
	"fmt"
	"github.com/47-11/spotifete/config"
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
	c.HTML(http.StatusOK, "index.html", gin.H{
		"time":               time.Now(),
		"activeSessionCount": service.ListeningSessionService().GetActiveSessionCount(),
		"totalSessionCount":  service.ListeningSessionService().GetTotalSessionCount(),
		"user":               user,
		"userSessions":       service.ListeningSessionService().GetActiveSessionsByOwnerId(*loginSession.UserId),
	})
}

func (TemplateController) NewListeningSession(c *gin.Context) {
	loginSession := service.LoginSessionService().GetSessionFromCookie(c)
	if loginSession == nil || loginSession.UserId == nil {
		c.Redirect(http.StatusSeeOther, "/spotify/login?redirectTo=/session/new")
		return
	}

	user := service.UserService().GetUserById(*loginSession.UserId)
	c.HTML(http.StatusOK, "newSession.html", gin.H{
		"user": user,
	})
}

func (TemplateController) NewListeningSessionSubmit(c *gin.Context) {
	loginSession := service.LoginSessionService().GetSessionFromCookie(c)
	if loginSession == nil || loginSession.UserId == nil {
		c.Redirect(http.StatusSeeOther, "/spotify/login?redirectTo=/session/new")
		return
	}

	user := service.UserService().GetUserById(*loginSession.UserId)
	if user == nil {
		c.String(http.StatusNotFound, "user not found")
		return
	}

	title := c.PostForm("title")
	if len(title) == 0 {
		c.String(http.StatusBadRequest, "title must not be empty")
		return
	}

	session, err := service.ListeningSessionService().NewSession(*user, title)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	c.Redirect(http.StatusSeeOther, fmt.Sprintf("/session/view/%s", *session.JoinId))
}

func (TemplateController) ViewSession(c *gin.Context) {
	joinId := c.Param("joinId")
	listeningSession := service.ListeningSessionService().GetSessionByJoinId(joinId)

	if listeningSession == nil {
		c.String(http.StatusNotFound, "session not found")
		return
	}

	ListeningSessionDto := service.ListeningSessionService().CreateDto(*listeningSession, true)

	queueLastUpdated := service.ListeningSessionService().GetQueueLastUpdated(*listeningSession).UTC().Format(time.RFC3339Nano)
	loginSession := service.LoginSessionService().GetSessionFromCookie(c)
	if loginSession == nil || loginSession.UserId == nil {
		c.HTML(http.StatusOK, "viewSession.html", gin.H{
			"queueLastUpdated": queueLastUpdated,
			"session":          ListeningSessionDto,
		})
		return
	}

	user := service.UserService().GetUserById(*loginSession.UserId)
	c.HTML(http.StatusOK, "viewSession.html", gin.H{
		"queueLastUpdated": queueLastUpdated,
		"session":          ListeningSessionDto,
		"user":             user,
	})

}

func (TemplateController) CloseListeningSession(c *gin.Context) {
	loginSession := service.LoginSessionService().GetSessionFromCookie(c)
	if loginSession == nil || loginSession.UserId == nil {
		c.Redirect(http.StatusUnauthorized, "/spotify/login")
		return
	}

	user := service.UserService().GetUserById(*loginSession.UserId)
	if user == nil {
		c.String(http.StatusNotFound, "user not found")
		return
	}

	joinId := c.PostForm("joinId")
	if len(joinId) == 0 {
		c.String(http.StatusBadRequest, "parameter joinId not present")
		return
	}

	loginSession := service.LoginSessionService().GetSessionFromCookie(c)
	if loginSession == nil || loginSession.UserId == nil {
		c.Redirect(http.StatusUnauthorized, fmt.Sprintf("/spotify/login?redirectTo=/session/view/%s", joinId))
		return
	}

	user := service.UserService().GetUserById(*loginSession.UserId)

	err := service.ListeningSessionService().CloseSession(*user, joinId)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	c.Redirect(http.StatusSeeOther, "/")
}

func (TemplateController) GetApp(c *gin.Context) {
	c.Redirect(http.StatusTemporaryRedirect, "/app/android")
}

func (TemplateController) GetAppAndroid(c *gin.Context) {
	c.Redirect(http.StatusTemporaryRedirect, config.GetConfig().GetString("spotifete.app.androidUrl"))
}

func (TemplateController) GetAppIOS(c *gin.Context) {
	c.String(http.StatusNotImplemented, "Sorry, the iOS app is not available yet!")
}
