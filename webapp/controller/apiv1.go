package controller

import (
	"github.com/47-11/spotifete/authentication"
	"github.com/47-11/spotifete/database/model"
	"github.com/47-11/spotifete/listeningSession"
	. "github.com/47-11/spotifete/model/webapp/api/v1"
	"github.com/47-11/spotifete/shared"
	"github.com/47-11/spotifete/users"
	"github.com/getsentry/sentry-go"
	"github.com/gin-gonic/gin"
	"github.com/google/logger"
	"net/http"
	"strconv"
	"strings"
)

type ApiV1Controller struct{ Controller }

func (controller ApiV1Controller) SetupWithBaseRouter(baseRouter *gin.Engine) {
	router := baseRouter.Group("/api/v1")

	router.GET("/", controller.Index)
	router.GET("/spotify/auth/new", authentication.ApiNewSession)
	router.GET("/spotify/auth/authenticated", authentication.ApiIsSessionAuthenticated)
	router.PATCH("/spotify/auth/invalidate", authentication.ApiInvalidateSession)
	router.GET("/spotify/auth/success", authentication.ApiCallbackSuccess)
	router.GET("/spotify/search/track", controller.SearchSpotifyTrack)
	router.GET("/spotify/search/playlist", controller.SearchSpotifyPlaylist)
	router.GET("/sessions/:joinId", listeningSession.ApiGetSession)
	router.DELETE("sessions/:joinId", controller.CloseListeningSession)
	router.POST("/sessions/:joinId/request", controller.RequestSong)
	router.GET("/sessions/:joinId/queuelastupdated", controller.QueueLastUpdated)
	router.GET("/sessions/:joinId/qrcode", controller.CreateQrCodeForListeningSession)
	router.POST("/sessions", listeningSession.ApiNewSession)
	router.GET("/users/:userId", controller.GetUser)
}

func (ApiV1Controller) Index(c *gin.Context) {
	c.String(http.StatusOK, "SpotiFete API v1")
}

func (controller ApiV1Controller) GetUser(c *gin.Context) {
	spotifyUserId := c.Param("userId")

	if spotifyUserId == "current" {
		controller.GetCurrentUser(c)
		return
	}

	fullUser := users.FindSimpleUser(model.SimpleUser{SpotifyId: spotifyUserId})

	if fullUser == nil {
		c.JSON(http.StatusNotFound, shared.ErrorResponse{Message: "user not found"})
	} else {
		c.JSON(http.StatusOK, fullUser)
	}
}

func (ApiV1Controller) GetCurrentUser(c *gin.Context) {
	loginSessionId := c.Query("sessionId")

	if len(loginSessionId) == 0 {
		c.JSON(http.StatusBadRequest, shared.ErrorResponse{Message: "session id not given"})
		return
	}

	loginSession := authentication.GetValidSession(loginSessionId)
	if loginSession == nil {
		c.JSON(http.StatusUnauthorized, shared.ErrorResponse{Message: "invalid login session"})
		return
	}

	if loginSession.UserId == nil {
		c.JSON(http.StatusUnauthorized, shared.ErrorResponse{Message: "not authenticated to spotify yet"})
		return
	}

	c.JSON(http.StatusOK, loginSession.User)
}

func (ApiV1Controller) SearchSpotifyTrack(c *gin.Context) {
	listeningSessionJoinId := c.Query("session")
	if len(listeningSessionJoinId) == 0 {
		c.JSON(http.StatusBadRequest, shared.ErrorResponse{Message: "session not specified"})
		return
	}

	query := c.Query("query")
	if len(query) == 0 {
		c.JSON(http.StatusBadRequest, shared.ErrorResponse{Message: "query not given"})
		return
	}

	limitPatameter := c.Query("limit")
	var limit int = -1
	if len(limitPatameter) > 0 {
		limitParsed, err := strconv.ParseInt(limitPatameter, 10, 0)
		if err != nil {
			c.JSON(http.StatusBadRequest, shared.ErrorResponse{Message: "invaid limit"})
			return
		}

		limit = int(limitParsed)
	} else {
		limit = 10
	}

	session := listeningSession.FindFullListeningSession(model.SimpleListeningSession{
		JoinId: &listeningSessionJoinId,
	})
	if session == nil {
		c.JSON(http.StatusNotFound, shared.ErrorResponse{Message: "session not found"})
		return
	}

	client := users.Client(session.Owner)

	tracks, spotifeteError := listeningSession.SearchTrack(*client, query, limit)
	if spotifeteError != nil {
		spotifeteError.SetJsonResponse(c)
		return
	}

	c.JSON(http.StatusOK, SearchTracksResponse{
		Query:   query,
		Results: tracks,
	})
}

func (ApiV1Controller) SearchSpotifyPlaylist(c *gin.Context) {
	listeningSessionJoinId := c.Query("session")
	if len(listeningSessionJoinId) == 0 {
		c.JSON(http.StatusBadRequest, shared.ErrorResponse{Message: "session not specified"})
		return
	}

	query := c.Query("query")
	if len(query) == 0 {
		c.JSON(http.StatusBadRequest, shared.ErrorResponse{Message: "query not given"})
		return
	}

	limitPatameter := c.Query("limit")
	var limit int
	if len(limitPatameter) > 0 {
		limitParsed, err := strconv.ParseInt(limitPatameter, 10, 0)
		if err != nil {
			c.JSON(http.StatusBadRequest, shared.ErrorResponse{Message: "invaid limit"})
			return
		}

		limit = int(limitParsed)
	} else {
		limit = 10
	}

	session := listeningSession.FindFullListeningSession(model.SimpleListeningSession{
		JoinId: &listeningSessionJoinId,
	})
	if session == nil {
		c.JSON(http.StatusNotFound, shared.ErrorResponse{Message: "session not found"})
		return
	}

	client := users.Client(session.Owner)

	playlists, spotifeteError := listeningSession.SearchPlaylist(*client, query, limit)
	if spotifeteError != nil {
		spotifeteError.SetJsonResponse(c)
		return
	}

	c.JSON(http.StatusOK, SearchPlaylistResponse{
		Query:   query,
		Results: playlists,
	})
}

func (ApiV1Controller) RequestSong(c *gin.Context) {
	requestBody := RequestSongRequest{}
	err := c.ShouldBindJSON(&requestBody)
	if err != nil {
		logger.Info("Invalid request body: " + err.Error())
		c.JSON(http.StatusBadRequest, shared.ErrorResponse{Message: "invalid requestBody body: " + err.Error()})
		return
	}

	sessionJoinId := c.Param("joinId")
	session := listeningSession.FindFullListeningSession(model.SimpleListeningSession{
		JoinId: &sessionJoinId,
	})
	if session == nil {
		c.JSON(http.StatusNotFound, shared.ErrorResponse{Message: "session not found"})
		return
	}
	if !session.Active {
		c.JSON(http.StatusBadRequest, shared.ErrorResponse{Message: "session is closed"})
		return
	}

	if listeningSession.IsTrackInQueue(session.SimpleListeningSession, requestBody.TrackId) {
		c.JSON(http.StatusBadRequest, shared.ErrorResponse{Message: "that song is already in the queue"})
		return
	}

	_, spotifeteError := listeningSession.RequestSong(*session, requestBody.TrackId)
	if spotifeteError == nil {
		c.Status(http.StatusNoContent)
	} else {
		spotifeteError.SetJsonResponse(c)
	}
}

func (ApiV1Controller) QueueLastUpdated(c *gin.Context) {
	sessionJoinId := c.Param("joinId")
	session := listeningSession.FindSimpleListeningSession(model.SimpleListeningSession{
		JoinId: &sessionJoinId,
	})
	if session == nil {
		c.JSON(http.StatusNotFound, shared.ErrorResponse{Message: "session not found"})
		return
	}

	c.JSON(http.StatusOK, QueueLastUpdatedResponse{QueueLastUpdated: listeningSession.GetQueueLastUpdated(*session)})
}

func (ApiV1Controller) CloseListeningSession(c *gin.Context) {
	sessionJoinId := c.Param("joinId")

	var request = CloseListeningSessionRequest{}
	err := c.BindJSON(&request)
	if err != nil {
		c.JSON(http.StatusBadRequest, shared.ErrorResponse{Message: "Invalid request body"})
		return
	}

	loginSessionId := request.LoginSessionId
	if loginSessionId == nil {
		c.JSON(http.StatusBadRequest, shared.ErrorResponse{Message: "SpotifyLogin session id not given"})
		return
	}

	loginSession := authentication.GetValidSession(*loginSessionId)
	if loginSession == nil {
		c.JSON(http.StatusUnauthorized, shared.ErrorResponse{Message: "Invalid login session"})
		return
	}

	if loginSession.User == nil {
		c.JSON(http.StatusUnauthorized, shared.ErrorResponse{Message: "Login session not authorized"})
		return
	}

	spotifeteError := listeningSession.CloseSession(*loginSession.User, sessionJoinId)
	if spotifeteError == nil {
		c.Status(http.StatusNoContent)
	} else {
		spotifeteError.SetJsonResponse(c)
	}
}

func (ApiV1Controller) CreateQrCodeForListeningSession(c *gin.Context) {
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

	qrCode, spotifeteError := listeningSession.GenerateQrCodeForSession(joinId, disableBorder)
	if spotifeteError != nil {
		spotifeteError.SetJsonResponse(c)
		return
	}

	qrCodeImageBytes, err := qrCode.PNG(size)
	if err != nil {
		sentry.CaptureException(err)
		logger.Error(err)
		c.JSON(http.StatusInternalServerError, shared.ErrorResponse{Message: err.Error()})
		return
	}

	c.Data(http.StatusOK, "image/png", qrCodeImageBytes)
}

func (ApiV1Controller) CallbackSuccess(c *gin.Context) {
	c.String(http.StatusOK, "Authentication successful. You can close this window now.")
}
