package listeningSession

import (
	"github.com/47-11/spotifete/authentication"
	"github.com/47-11/spotifete/database/model"
	"github.com/47-11/spotifete/listeningSession/model/api"
	"github.com/47-11/spotifete/shared"
	"github.com/gin-gonic/gin"
	"net/http"
)

func ApiNewSession(c *gin.Context) {
	request := api.NewSessionRequest{}
	err := c.ShouldBindJSON(&request)
	if err != nil {
		c.JSON(http.StatusBadRequest, shared.ErrorResponse{Message: "invalid requestBody: " + err.Error()})
		return
	}

	spotifeteError := request.Validate()
	if spotifeteError != nil {
		spotifeteError.SetJsonResponse(c)
		return
	}

	loginSession := authentication.GetValidSession(request.LoginSessionId)
	if loginSession == nil {
		c.JSON(http.StatusUnauthorized, shared.ErrorResponse{Message: "invalid login session"})
		return
	}

	if loginSession.User == nil {
		c.JSON(http.StatusUnauthorized, shared.ErrorResponse{Message: "login session without user"})
		return
	}

	createdSession, spotifeteError := NewSession(*loginSession.User, request.ListeningSessionTitle)
	if spotifeteError != nil {
		spotifeteError.SetJsonResponse(c)
		return
	}

	c.JSON(http.StatusOK, createdSession)
}

func ApiGetSession(c *gin.Context) {
	joinId := c.Param("joinId")
	session := FindFullListeningSession(model.SimpleListeningSession{
		JoinId: &joinId,
	})

	if session == nil {
		c.JSON(http.StatusNotFound, shared.ErrorResponse{Message: "Session not found."})
	} else {
		c.JSON(http.StatusOK, session)
	}
}

func ApiCloseSession(c *gin.Context) {
	request := shared.AuthenticatedRequest{}
	err := c.ShouldBindJSON(&request)
	if err != nil {
		c.JSON(http.StatusBadRequest, shared.ErrorResponse{Message: "invalid requestBody: " + err.Error()})
		return
	}

	spotifeteError := request.Validate()
	if spotifeteError != nil {
		spotifeteError.SetJsonResponse(c)
		return
	}

	loginSession := authentication.GetValidSession(request.LoginSessionId)
	if loginSession == nil {
		c.JSON(http.StatusUnauthorized, shared.ErrorResponse{Message: "invalid login session"})
		return
	}

	if loginSession.User == nil {
		c.JSON(http.StatusUnauthorized, shared.ErrorResponse{Message: "login session without user"})
		return
	}

	CloseSession(*loginSession.User, request.LoginSessionId)
}
