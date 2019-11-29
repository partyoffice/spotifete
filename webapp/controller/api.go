package controller

import (
	"github.com/47-11/spotifete/model"
	"github.com/47-11/spotifete/service"
	. "github.com/47-11/spotifete/webapp/model/api/v1"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

type ApiController struct {
	sessionService service.ListeningSessionService
	userService    service.UserService
	spotifyService service.SpotifyService
}

func (controller ApiController) Index(c *gin.Context) {
	c.String(http.StatusOK, "SpotiFete API v1")
}

func (controller ApiController) GetSession(c *gin.Context) {
	sessionId, err := strconv.ParseInt(c.Param("sessionId"), 0, 0)
	session, err := controller.sessionService.GetSessionByJoinId(uint(sessionId))

	if err != nil {
		if _, ok := err.(model.EntryNotFoundError); ok {
			c.String(http.StatusNotFound, err.Error())
		} else {
			c.String(http.StatusInternalServerError, err.Error())
		}
	} else {
		c.JSON(http.StatusOK, session)
	}
}

func (controller ApiController) GetUser(c *gin.Context) {
	userId, err := strconv.ParseInt(c.Param("userId"), 0, 0)
	user, err := controller.userService.GetUserById(uint(userId))

	if err != nil {
		if _, notFound := err.(model.EntryNotFoundError); notFound {
			c.String(http.StatusNotFound, err.Error())
		} else {
			c.String(http.StatusInternalServerError, err.Error())
		}
	} else {
		c.JSON(http.StatusOK, user)
	}
}

func (controller ApiController) GetAuthUrl(c *gin.Context) {
	url, sessionId := controller.spotifyService.NewAuthUrl()
	c.JSON(http.StatusOK, model.AuthUrlDto{
		Url:       url,
		SessionId: sessionId,
	})
}

func (controller ApiController) DidAuthSucceed(c *gin.Context) {
	sessionId := c.Query("sessionId")
	if sessionId == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{Message: "Parameter sessionId not found."})
		return
	}

	session := service.LoginSessionService().GetSessionBySessionId(sessionId)
	if session == nil || session.UserId == nil || !service.LoginSessionService().IsSessionValid(*session) {
		c.JSON(http.StatusUnauthorized, struct {
			Authenticated bool `json:"authenticated"`
		}{
			Authenticated: false,
		})
		return
	} else {
		c.JSON(http.StatusOK, struct {
			Authenticated bool `json:"authenticated"`
		}{
			Authenticated: true,
		})
	}
}

func (controller ApiController) InvalidateSessionId(c *gin.Context) {
	requestBody := struct {
		SessionId string `json:"sessionId"`
	}{}

	err := c.ShouldBindJSON(&requestBody)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Message: "Session id not given."})
		return
	}

	err = service.LoginSessionService().InvalidateSessionBySessionId(requestBody.SessionId)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Message: err.Error()})
		return
	}

	c.Status(http.StatusOK)
}
