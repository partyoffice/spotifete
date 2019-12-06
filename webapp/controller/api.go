package controller

import (
	. "github.com/47-11/spotifete/model/dto"
	. "github.com/47-11/spotifete/model/webapp/api/v1"
	"github.com/47-11/spotifete/service"
	"github.com/gin-gonic/gin"
	"net/http"
)

type ApiController struct{}

func (ApiController) Index(c *gin.Context) {
	c.String(http.StatusOK, "SpotiFete API v1")
}

func (ApiController) GetSession(c *gin.Context) {
	sessionId := c.Param("sessionId")
	session := service.ListeningSessionService().GetSessionByJoinId(sessionId)

	if session == nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Message: "session not found"})
	} else {
		c.JSON(http.StatusOK, ListeningSessionDto{}.FromDatabaseModel(*session))
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
		c.JSON(http.StatusNotFound, ErrorResponse{Message: "user not found"})
	} else {
		c.JSON(http.StatusOK, service.UserService().CreateDtoWithAdditionalInformation(user))
	}
}

func (ApiController) GetCurrentUser(c *gin.Context) {
	loginSessionId := c.Query("sessionId")

	if len(loginSessionId) == 0 {
		c.JSON(http.StatusBadRequest, ErrorResponse{Message: "session id not given"})
		return
	}

	loginSession := service.LoginSessionService().GetSessionBySessionId(loginSessionId)
	if loginSession == nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Message: "unknown session id"})
		return
	}

	user := service.UserService().GetUserById(*loginSession.UserId)
	c.JSON(http.StatusOK, service.UserService().CreateDtoWithAdditionalInformation(user))
}

func (ApiController) GetAuthUrl(c *gin.Context) {
	url, sessionId := service.SpotifyService().NewAuthUrl()
	c.JSON(http.StatusOK, GetAuthUrlResponse{
		Url:       url,
		SessionId: sessionId,
	})
}

func (controller ApiController) DidAuthSucceed(c *gin.Context) {
	sessionId := c.Query("sessionId")
	if sessionId == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{Message: "parameter sessionId not found."})
		return
	}

	session := service.LoginSessionService().GetSessionBySessionId(sessionId)
	if session == nil || session.UserId == nil || !service.LoginSessionService().IsSessionValid(*session) {
		c.JSON(http.StatusUnauthorized, DidAuthSucceedResponse{Authenticated: false})
		return
	} else {
		c.JSON(http.StatusOK, DidAuthSucceedResponse{Authenticated: true})
	}
}

func (controller ApiController) InvalidateSessionId(c *gin.Context) {
	var requestBody InvalidateSessionIdRequest

	err := c.ShouldBindJSON(&requestBody)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Message: "session id not given."})
		return
	}

	err = service.LoginSessionService().InvalidateSessionBySessionId(requestBody.SessionId)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Message: err.Error()})
		return
	}

	c.Status(http.StatusOK)
}

func (controller ApiController) SearchSpotifyTrack(c *gin.Context) {
	listeningSessionJoinId := c.Query("session")
	if len(listeningSessionJoinId) == 0 {
		c.JSON(http.StatusBadRequest, ErrorResponse{Message: "session not specified"})
		return
	}

	query := c.Query("query")
	if len(query) == 0 {
		c.JSON(http.StatusBadRequest, ErrorResponse{Message: "query not given"})
		return
	}

	session := service.ListeningSessionService().GetSessionByJoinId(listeningSessionJoinId)
	if session == nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Message: "session not found"})
		return
	}

	// TODO: Cache spotify clients (#7)
	user := service.UserService().GetUserById(session.OwnerId)
	token := user.GetToken()
	client := service.SpotifyService().GetAuthenticator().NewClient(token)

	tracks, err := service.SpotifyService().SearchTrack(&client, query, 5)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, SearchTracksResponse{
		Query:   query,
		Results: tracks,
	})
}
