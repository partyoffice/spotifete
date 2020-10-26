package listeningSession

import (
	"github.com/47-11/spotifete/database/model"
	"github.com/47-11/spotifete/listeningSession"
	"github.com/47-11/spotifete/webapp/apiv2/shared"
	"github.com/gin-gonic/gin"
	"net/http"
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
