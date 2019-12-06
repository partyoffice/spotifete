package controller

import (
	"github.com/47-11/spotifete/model"
	"github.com/47-11/spotifete/service"
	"github.com/47-11/spotifete/webapp/model/api/v1/dto"
	"github.com/47-11/spotifete/webapp/model/api/v1/shared"
	"github.com/gin-gonic/gin"
	"net/http"
)

type ApiController struct{}

func (controller ApiController) Index(c *gin.Context) {
	c.String(http.StatusOK, "SpotiFete API v1")
}

func (controller ApiController) GetSession(c *gin.Context) {
	sessionId := c.Param("sessionId")
	session := service.ListeningSessionService().GetSessionByJoinId(sessionId)

	if session == nil {
		c.JSON(http.StatusNotFound, shared.ErrorResponse{Message: "session not found"})
	} else {
		c.JSON(http.StatusOK, dto.ListeningSessionDto{}.FromDatabaseModel(*session))
	}
}

func (controller ApiController) GetUser(c *gin.Context) {
	userId := c.Param("userId")
	if userId == "current" {
		controller.GetCurrentUser(c)
		return
	}

	user := service.UserService().GetUserBySpotifyId(userId)

	if user == nil {
		c.JSON(http.StatusNotFound, shared.ErrorResponse{Message: "user not found"})
	} else {
		c.JSON(http.StatusOK, service.UserService().CreateDtoWithAdditionalInformation(user))
	}
}

func (controller ApiController) GetCurrentUser(c *gin.Context) {
	loginSessionId := c.Query("sessionId")

	if len(loginSessionId) == 0 {
		c.JSON(http.StatusBadRequest, shared.ErrorResponse{Message: "session id not given"})
		return
	}

	loginSession := service.LoginSessionService().GetSessionBySessionId(loginSessionId)
	if loginSession == nil {
		c.JSON(http.StatusUnauthorized, shared.ErrorResponse{Message: "unknown session id"})
		return
	}

	user := service.UserService().GetUserById(*loginSession.UserId)
	c.JSON(http.StatusOK, service.UserService().CreateDtoWithAdditionalInformation(user))
}

func (controller ApiController) GetAuthUrl(c *gin.Context) {
	url, sessionId := service.SpotifyService().NewAuthUrl()
	c.JSON(http.StatusOK, model.AuthUrlDto{
		Url:       url,
		SessionId: sessionId,
	})
}

func (controller ApiController) DidAuthSucceed(c *gin.Context) {
	sessionId := c.Query("sessionId")
	if sessionId == "" {
		c.JSON(http.StatusBadRequest, shared.ErrorResponse{Message: "Parameter sessionId not found."})
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
		c.JSON(http.StatusBadRequest, shared.ErrorResponse{Message: "Session id not given."})
		return
	}

	err = service.LoginSessionService().InvalidateSessionBySessionId(requestBody.SessionId)
	if err != nil {
		c.JSON(http.StatusBadRequest, shared.ErrorResponse{Message: err.Error()})
		return
	}

	c.Status(http.StatusOK)
}
