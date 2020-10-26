package listeningSession

import (
	"github.com/47-11/spotifete/database/model"
	"github.com/47-11/spotifete/listeningSession"
	. "github.com/47-11/spotifete/webapp/apiv2/shared"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

func newSession(c *gin.Context) {
	request := NewSessionRequest{}
	err := c.ShouldBindJSON(&request)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Message: "invalid requestBody: " + err.Error()})
		return
	}

	spotifeteError := request.Validate()
	if spotifeteError != nil {
		SetJsonError(*spotifeteError, c)
		return
	}

	authenticatedUser, spotifeteError := request.GetUser()
	if spotifeteError != nil {
		SetJsonError(*spotifeteError, c)
		return
	}

	createdSession, spotifeteError := listeningSession.NewSession(authenticatedUser, request.ListeningSessionTitle)
	if spotifeteError != nil {
		SetJsonError(*spotifeteError, c)
		return
	}

	c.JSON(http.StatusOK, createdSession)
}

func getSession(c *gin.Context) {
	joinId := c.Param("joinId")
	session := listeningSession.FindFullListeningSession(model.SimpleListeningSession{
		JoinId: &joinId,
	})

	if session == nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Message: "Session not found."})
	} else {
		c.JSON(http.StatusOK, session)
	}
}

func closeSession(c *gin.Context) {
	request := AuthenticatedRequest{}
	err := c.ShouldBindJSON(&request)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Message: "invalid requestBody: " + err.Error()})
		return
	}

	authenticatedUser, spotifeteError := request.GetUser()
	if spotifeteError != nil {
		SetJsonError(*spotifeteError, c)
		return
	}

	joinId := c.Param("joinId")
	spotifeteError = listeningSession.CloseSession(authenticatedUser, joinId)
	if spotifeteError == nil {
		c.Status(http.StatusNoContent)
	} else {
		SetJsonError(*spotifeteError, c)
	}
}

func getSessionQueue(c *gin.Context) {
	joinId := c.Param("joinId")
	session := listeningSession.FindSimpleListeningSession(model.SimpleListeningSession{
		JoinId: &joinId,
	})
	if session == nil {
		c.JSON(http.StatusNotFound, "Session not found")
		return
	}

	currentlyPlayingRequest := listeningSession.GetCurrentlyPlayingRequest(*session)
	upNextRequest := listeningSession.GetUpNextRequest(*session)
	queue := listeningSession.GetSessionQueueInDemocraticOrder(*session)
	c.JSON(http.StatusOK, GetSessionQueueResponse{
		CurrentlyPlayingRequest: currentlyPlayingRequest,
		UpNextRequest:           upNextRequest,
		Queue:                   queue,
	})
}

func queueLastUpdated(c *gin.Context) {
	joinId := c.Param("joinId")
	session := listeningSession.FindSimpleListeningSession(model.SimpleListeningSession{
		JoinId: &joinId,
	})
	if session == nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Message: "Session not found."})
		return
	}

	queueLastUpdated := listeningSession.GetQueueLastUpdated(*session)
	c.JSON(http.StatusOK, QueueLastUpdatedResponse{QueueLastUpdated: queueLastUpdated})
}

func qrCode(c *gin.Context) {
	joinId := c.Param("joinId")
	disableBorder := "true" == c.Query("disableBorder")

	qrCode, spotifeteError := listeningSession.QrCodeAsPng(joinId, disableBorder, 512)
	if spotifeteError != nil {
		SetJsonError(*spotifeteError, c)
		return
	}

	c.Data(http.StatusOK, "image/png", qrCode.Bytes())
}

func searchTrack(c *gin.Context) {
	joinId := c.Param("joinId")
	session := listeningSession.FindFullListeningSession(model.SimpleListeningSession{
		JoinId: &joinId,
	})
	if session == nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Message: "Session not found."})
		return
	}

	query := c.Query("query")
	if len(query) == 0 {
		c.JSON(http.StatusBadRequest, ErrorResponse{Message: "Missing parameter query."})
		return
	}

	limitParameter := c.Query("limit")
	var limit = 20
	if len(limitParameter) > 0 {
		parsedLimit, err := strconv.ParseUint(limitParameter, 10, 0)
		if err == nil {
			limit = int(parsedLimit)
		} else {
			c.JSON(http.StatusBadRequest, ErrorResponse{Message: "Invalid limit."})
			return
		}
	}

	tracks, spotifeteError := listeningSession.SearchTrack(*session, query, limit)
	if spotifeteError == nil {
		c.JSON(http.StatusOK, SearchTracksResponse{
			Query:  query,
			Tracks: tracks,
		})
	} else {
		SetJsonError(*spotifeteError, c)
	}
}

func searchPlaylist(c *gin.Context) {
	joinId := c.Param("joinId")
	session := listeningSession.FindFullListeningSession(model.SimpleListeningSession{
		JoinId: &joinId,
	})
	if session == nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Message: "Session not found."})
		return
	}

	query := c.Query("query")
	if len(query) == 0 {
		c.JSON(http.StatusBadRequest, ErrorResponse{Message: "Missing parameter query."})
		return
	}

	limitParameter := c.Query("limit")
	var limit = 20
	if len(limitParameter) > 0 {
		parsedLimit, err := strconv.ParseUint(limitParameter, 10, 0)
		if err == nil {
			limit = int(parsedLimit)
		} else {
			c.JSON(http.StatusBadRequest, ErrorResponse{Message: "Invalid limit."})
			return
		}
	}

	playlists, spotifeteError := listeningSession.SearchPlaylist(*session, query, limit)
	if spotifeteError == nil {
		c.JSON(http.StatusOK, SearchPlaylistResponse{
			Query:     query,
			Playlists: playlists,
		})
	} else {
		SetJsonError(*spotifeteError, c)
	}
}

func requestTrack(c *gin.Context) {
	request := RequestTrackRequest{}
	err := c.ShouldBindJSON(&request)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Message: "invalid requestBody: " + err.Error()})
		return
	}

	spotifeteError := request.Validate()
	if spotifeteError != nil {
		SetJsonError(*spotifeteError, c)
		return
	}

	joinId := c.Param("joinId")
	session := listeningSession.FindFullListeningSession(model.SimpleListeningSession{
		JoinId: &joinId,
	})
	if session == nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Message: "Listening session not found."})
		return
	}

	_, spotifeteError = listeningSession.RequestSong(*session, request.TrackId)
	if spotifeteError == nil {
		c.Status(http.StatusNoContent)
	} else {
		SetJsonError(*spotifeteError, c)
	}
}

func changeFallbackPlaylist(c *gin.Context) {
	request := ChangeFallbackPlaylistRequest{}
	err := c.ShouldBindJSON(&request)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Message: "invalid requestBody: " + err.Error()})
		return
	}

	spotifeteError := request.Validate()
	if spotifeteError != nil {
		SetJsonError(*spotifeteError, c)
		return
	}

	authenticatedUser, spotifeteError := request.GetUser()
	if spotifeteError != nil {
		SetJsonError(*spotifeteError, c)
		return
	}

	joinId := c.Param("joinId")
	session := listeningSession.FindSimpleListeningSession(model.SimpleListeningSession{
		JoinId: &joinId,
	})
	if session == nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Message: "Listening session not found."})
		return
	}

	spotifeteError = listeningSession.ChangeFallbackPlaylist(*session, authenticatedUser, request.NewFallbackPlaylistId)
	if spotifeteError == nil {
		c.Status(http.StatusNoContent)
	} else {
		SetJsonError(*spotifeteError, c)
	}
}

func removeFallbackPlaylist(c *gin.Context) {
	request := AuthenticatedRequest{}
	err := c.ShouldBindJSON(&request)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Message: "invalid requestBody: " + err.Error()})
		return
	}

	authenticatedUser, spotifeteError := request.GetUser()
	if spotifeteError != nil {
		SetJsonError(*spotifeteError, c)
		return
	}

	joinId := c.Param("joinId")
	session := listeningSession.FindSimpleListeningSession(model.SimpleListeningSession{
		JoinId: &joinId,
	})
	if session == nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Message: "Listening session not found."})
		return
	}

	spotifeteError = listeningSession.RemoveFallbackPlaylist(*session, authenticatedUser)
	if spotifeteError == nil {
		c.Status(http.StatusNoContent)
	} else {
		SetJsonError(*spotifeteError, c)
	}
}
