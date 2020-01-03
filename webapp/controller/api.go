package controller

import (
	. "github.com/47-11/spotifete/model/webapp/api/v1"
	"github.com/47-11/spotifete/service"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

type ApiController struct{}

func (ApiController) Index(c *gin.Context) {
	c.String(http.StatusOK, "SpotiFete API v1")
}

func (ApiController) GetSession(c *gin.Context) {
	sessionJoinId := c.Param("joinId")

	session := service.ListeningSessionService().GetSessionByJoinId(sessionJoinId)
	if session == nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Message: "session not found"})
	} else {
		c.JSON(http.StatusOK, service.ListeningSessionService().CreateDto(*session, true))
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
		c.JSON(http.StatusOK, service.UserService().CreateDto(*user, true))
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
		c.JSON(http.StatusBadRequest, ErrorResponse{Message: "unknown session id"})
		return
	}

	user := service.UserService().GetUserById(*loginSession.UserId)
	c.JSON(http.StatusOK, service.UserService().CreateDto(*user, true))
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

	err = service.LoginSessionService().InvalidateSessionBySessionId(requestBody.LoginSessionId)
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

	limitPatameter := c.Query("limit")
	var limit int = -1
	if len(limitPatameter) > 0 {
		limitParsed, err := strconv.ParseInt(limitPatameter, 10, 0)
		if err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{Message: "could not parse limit"})
			return
		}

		limit = int(limitParsed)
	} else {
		limit = 10
	}

	session := service.ListeningSessionService().GetSessionByJoinId(listeningSessionJoinId)
	if session == nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Message: "session not found"})
		return
	}

	user := service.UserService().GetUserById(session.OwnerId)
	client := service.SpotifyService().GetClientForUser(*user)

	tracks, err := service.SpotifyService().SearchTrack(*client, query, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, SearchTracksResponse{
		Query:   query,
		Results: tracks,
	})
}

func (controller ApiController) RequestSong(c *gin.Context) {
	requestBody := RequestSongRequest{}
	err := c.ShouldBindJSON(&requestBody)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Message: "invalid requestBody body"})
		return
	}

	sessionJoinId := c.Param("joinId")
	session := service.ListeningSessionService().GetSessionByJoinId(sessionJoinId)
	if session == nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Message: "session not found"})
		return
	}
	if !session.Active {
		c.JSON(http.StatusBadRequest, ErrorResponse{Message: "session is closed"})
		return
	}

	err = service.ListeningSessionService().RequestSong(session, requestBody.TrackId)
	if err == nil {
		c.Status(http.StatusOK)
	} else {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Message: err.Error()})
	}
}

func (controller ApiController) QueueLastUpdated(c *gin.Context) {
	sessionJoinId := c.Param("joinId")
	session := service.ListeningSessionService().GetSessionByJoinId(sessionJoinId)
	if session == nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Message: "session not found"})
		return
	}

	c.JSON(http.StatusOK, QueueLastUpdatedResponse{QueueLastUpdated: service.ListeningSessionService().GetQueueLastUpdated(*session)})
}

func (controller ApiController) CreateListeningSession(c *gin.Context) {
	requestBody := CreateListeningSessionRequest{}
	err := c.ShouldBindJSON(&requestBody)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Message: "invalid request body"})
		return
	}

	if requestBody.LoginSessionId == nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Message: "required parameter loginSessionId not present"})
		return
	}

	if requestBody.ListeningSessionTitle == nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Message: "required parameter listeningSessionTitle not present"})
		return
	}

	loginSession := service.LoginSessionService().GetSessionBySessionId(*requestBody.LoginSessionId)
	if loginSession == nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Message: "invalid login session"})
		return
	}

	if loginSession.UserId == nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Message: "login session is not authorized to spotify yet"})
		return
	}

	owner := service.UserService().GetUserById(*loginSession.UserId)
	createdSession, err := service.ListeningSessionService().NewSession(*owner, *requestBody.ListeningSessionTitle)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, service.ListeningSessionService().CreateDto(*createdSession, true))
}
