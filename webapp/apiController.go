package webapp

import (
	. "github.com/47-11/spotifete/model/webapp/api/v1"
	"github.com/47-11/spotifete/service"
	"github.com/getsentry/sentry-go"
	"github.com/gin-gonic/gin"
	"github.com/google/logger"
	"net/http"
	"strconv"
	"strings"
)

type ApiController struct{}

func (c ApiController) SetupRoutes(baseRouter *gin.Engine) {
	router := baseRouter.Group("/api/v1")

	router.GET("/", c.Index)
	router.GET("/spotify/auth/new", c.GetAuthUrl)
	router.GET("/spotify/auth/authenticated", c.DidAuthSucceed)
	router.PATCH("/spotify/auth/invalidate", c.InvalidateSessionId)
	router.GET("/spotify/search/track", c.SearchSpotifyTrack)
	router.GET("/spotify/search/playlist", c.SearchSpotifyPlaylist)
	router.GET("/sessions/:joinId", c.GetSession)
	router.DELETE("sessions/:joinId", c.CloseListeningSession)
	router.POST("/sessions/:joinId/request", c.RequestSong)
	router.GET("/sessions/:joinId/queuelastupdated", c.QueueLastUpdated)
	router.GET("/sessions/:joinId/qrcode", c.CreateQrCodeForListeningSession)
	router.POST("/sessions", c.CreateListeningSession)
	router.GET("/users/:userId", c.GetUser)
}

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

func (ApiController) GetUser(c *gin.Context) {
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

	loginSession := service.LoginSessionService().GetSessionBySessionId(loginSessionId, true)
	if loginSession == nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Message: "invalid login session"})
		return
	}

	if loginSession.UserId == nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Message: "not authenticated to spotify yet"})
		return
	}

	user := service.UserService().GetUserById(*loginSession.UserId)
	c.JSON(http.StatusOK, service.UserService().CreateDto(*user, true))
}

func (ApiController) GetAuthUrl(c *gin.Context) {
	url, sessionId := service.SpotifyService().NewAuthUrl("/spotify/apicallback")
	c.JSON(http.StatusOK, GetAuthUrlResponse{
		Url:       url,
		SessionId: sessionId,
	})
}

func (ApiController) DidAuthSucceed(c *gin.Context) {
	sessionId := c.Query("sessionId")
	if sessionId == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{Message: "session id not given"})
		return
	}

	session := service.LoginSessionService().GetSessionBySessionId(sessionId, false)
	if session == nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Message: "session not found"})
		return
	}

	if service.LoginSessionService().IsSessionValid(*session) && session.UserId != nil {
		c.JSON(http.StatusOK, DidAuthSucceedResponse{Authenticated: true})
	} else {
		c.JSON(http.StatusUnauthorized, DidAuthSucceedResponse{Authenticated: false})
	}
}

func (ApiController) InvalidateSessionId(c *gin.Context) {
	var requestBody InvalidateSessionIdRequest

	err := c.ShouldBindJSON(&requestBody)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Message: "session id not given"})
		return
	}

	err = service.LoginSessionService().InvalidateSessionBySessionId(requestBody.LoginSessionId)
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Message: "session not found"})
		return
	}

	c.Status(http.StatusNoContent)
}

func (ApiController) SearchSpotifyTrack(c *gin.Context) {
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
			c.JSON(http.StatusBadRequest, ErrorResponse{Message: "invaid limit"})
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
		sentry.CaptureException(err)
		logger.Error(err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, SearchTracksResponse{
		Query:   query,
		Results: tracks,
	})
}

func (ApiController) SearchSpotifyPlaylist(c *gin.Context) {
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
	var limit int
	if len(limitPatameter) > 0 {
		limitParsed, err := strconv.ParseInt(limitPatameter, 10, 0)
		if err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{Message: "invaid limit"})
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

	playlists, err := service.SpotifyService().SearchPlaylist(*client, query, limit)
	if err != nil {
		sentry.CaptureException(err)
		logger.Error(err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, SearchPlaylistResponse{
		Query:   query,
		Results: playlists,
	})
}

func (ApiController) RequestSong(c *gin.Context) {
	requestBody := RequestSongRequest{}
	err := c.ShouldBindJSON(&requestBody)
	if err != nil {
		logger.Info("Invalid request body: " + err.Error())
		c.JSON(http.StatusBadRequest, ErrorResponse{Message: "invalid requestBody body: " + err.Error()})
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

	if service.ListeningSessionService().IsTrackInQueue(*session, requestBody.TrackId) {
		c.JSON(http.StatusBadRequest, ErrorResponse{Message: "that song is already in the queue"})
		return
	}

	err = service.ListeningSessionService().RequestSong(*session, requestBody.TrackId)
	if err == nil {
		c.Status(http.StatusNoContent)
	} else {
		sentry.CaptureException(err)
		logger.Error(err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{Message: err.Error()})
	}
}

func (ApiController) QueueLastUpdated(c *gin.Context) {
	sessionJoinId := c.Param("joinId")
	session := service.ListeningSessionService().GetSessionByJoinId(sessionJoinId)
	if session == nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Message: "session not found"})
		return
	}

	c.JSON(http.StatusOK, QueueLastUpdatedResponse{QueueLastUpdated: service.ListeningSessionService().GetQueueLastUpdated(*session)})
}

func (ApiController) CreateListeningSession(c *gin.Context) {
	requestBody := CreateListeningSessionRequest{}
	err := c.ShouldBindJSON(&requestBody)
	if err != nil {
		logger.Info("Invalid request body: " + err.Error())
		c.JSON(http.StatusBadRequest, ErrorResponse{Message: "invalid requestBody body: " + err.Error()})
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

	loginSession := service.LoginSessionService().GetSessionBySessionId(*requestBody.LoginSessionId, true)
	if loginSession == nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Message: "invalid login session"})
		return
	}

	if loginSession.UserId == nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Message: "not authenticated to spotify yet"})
		return
	}

	owner := service.UserService().GetUserById(*loginSession.UserId)
	createdSession, err := service.ListeningSessionService().NewSession(*owner, *requestBody.ListeningSessionTitle)
	if err != nil {
		sentry.CaptureException(err)
		logger.Error(err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, service.ListeningSessionService().CreateDto(*createdSession, true))
}

func (ApiController) CloseListeningSession(c *gin.Context) {
	sessionJoinId := c.Param("joinId")

	var request = CloseListeningSessionRequest{}
	err := c.BindJSON(&request)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Message: "Invalid request body"})
		return
	}

	loginSessionId := request.LoginSessionId
	if loginSessionId == nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Message: "SpotifyLogin session id not given"})
		return
	}

	loginSession := service.LoginSessionService().GetSessionBySessionId(*loginSessionId, true)
	if loginSession == nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Message: "Invalid login session"})
		return
	}

	user := service.UserService().GetUserById(*loginSession.UserId)

	err = service.ListeningSessionService().CloseSession(*user, sessionJoinId)
	if err == nil {
		c.Status(http.StatusNoContent)
	} else {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Message: err.Error()})
	}
}

func (ApiController) CreateQrCodeForListeningSession(c *gin.Context) {
	joinId := c.Param("joinId")
	disableBorder := strings.EqualFold("true", c.Query("disableBorder"))
	sizeOverride := c.Query("size")

	size := 512
	if len(sizeOverride) > 0 {
		parsed, err := strconv.Atoi(sizeOverride)
		if err != nil || parsed <= 0 {
			c.JSON(http.StatusBadRequest, "Invalid size")
			return
		}

		size = parsed
	}

	qrCode, err := service.ListeningSessionService().GenerateQrCodeForSession(joinId, disableBorder)
	if err != nil {
		sentry.CaptureException(err)
		logger.Error(err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{Message: err.Error()})
		return
	}

	qrCodeImageBytes, err := qrCode.PNG(size)
	if err != nil {
		sentry.CaptureException(err)
		logger.Error(err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{Message: err.Error()})
		return
	}

	c.Data(http.StatusOK, "image/png", qrCodeImageBytes)
}
