package listeningSession

import (
	"github.com/47-11/spotifete/database/model"
	"github.com/47-11/spotifete/listeningSession"
	"github.com/47-11/spotifete/webapp/apiv2/shared"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

func newSession(c *gin.Context) {
	request := NewSessionRequest{}
	err := c.ShouldBindJSON(&request)
	if err != nil {
		c.JSON(http.StatusBadRequest, shared.ErrorResponse{Message: "invalid requestBody: " + err.Error()})
		return
	}

	spotifeteError := request.Validate()
	if spotifeteError != nil {
		shared.SetJsonError(*spotifeteError, c)
		return
	}

	authenticatedUser, spotifeteError := request.GetUser()
	if spotifeteError != nil {
		shared.SetJsonError(*spotifeteError, c)
		return
	}

	createdSession, spotifeteError := listeningSession.NewSession(authenticatedUser, request.ListeningSessionTitle)
	if spotifeteError != nil {
		shared.SetJsonError(*spotifeteError, c)
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
		c.JSON(http.StatusNotFound, shared.ErrorResponse{Message: "Session not found."})
	} else {
		c.JSON(http.StatusOK, session)
	}
}

func closeSession(c *gin.Context) {
	request := shared.AuthenticatedRequest{}
	err := c.ShouldBindJSON(&request)
	if err != nil {
		c.JSON(http.StatusBadRequest, shared.ErrorResponse{Message: "invalid requestBody: " + err.Error()})
		return
	}

	authenticatedUser, spotifeteError := request.GetUser()
	if spotifeteError != nil {
		shared.SetJsonError(*spotifeteError, c)
		return
	}

	joinId := c.Param("joinId")
	spotifeteError = listeningSession.CloseSession(authenticatedUser, joinId)
	if spotifeteError == nil {
		c.Status(http.StatusNoContent)
	} else {
		shared.SetJsonError(*spotifeteError, c)
	}
}

func searchTrack(c *gin.Context) {
	joinId := c.Param("joinId")
	session := listeningSession.FindFullListeningSession(model.SimpleListeningSession{
		JoinId: &joinId,
	})
	if session == nil {
		c.JSON(http.StatusNotFound, shared.ErrorResponse{Message: "Session not found."})
		return
	}

	query := c.Query("query")
	if len(query) == 0 {
		c.JSON(http.StatusBadRequest, shared.ErrorResponse{Message: "Missing parameter query."})
		return
	}

	limitParameter := c.Query("limit")
	var limit = 20
	if len(limitParameter) > 0 {
		parsedLimit, err := strconv.ParseUint(limitParameter, 10, 0)
		if err == nil {
			limit = int(parsedLimit)
		} else {
			c.JSON(http.StatusBadRequest, shared.ErrorResponse{Message: "Invalid limit."})
			return
		}
	}

	listeningSession.SearchTrack(*session, query, limit)
}

func searchPlaylist(c *gin.Context) {
	joinId := c.Param("joinId")
	session := listeningSession.FindFullListeningSession(model.SimpleListeningSession{
		JoinId: &joinId,
	})
	if session == nil {
		c.JSON(http.StatusNotFound, shared.ErrorResponse{Message: "Session not found."})
		return
	}

	query := c.Query("query")
	if len(query) == 0 {
		c.JSON(http.StatusBadRequest, shared.ErrorResponse{Message: "Missing parameter query."})
		return
	}

	limitParameter := c.Query("limit")
	var limit = 20
	if len(limitParameter) > 0 {
		parsedLimit, err := strconv.ParseUint(limitParameter, 10, 0)
		if err == nil {
			limit = int(parsedLimit)
		} else {
			c.JSON(http.StatusBadRequest, shared.ErrorResponse{Message: "Invalid limit."})
			return
		}
	}

	listeningSession.SearchPlaylist(*session, query, limit)
}
