package controller

import (
	"github.com/47-11/spotifete/service"
	"github.com/gin-gonic/gin"
	"net/http"
)

type ApiController struct{}

var sessionService = new(service.SessionService)

func (a ApiController) Index(c *gin.Context) {
	c.String(http.StatusOK, "SpotiFete API v1")
}

func (a ApiController) GetActiveSessions(c *gin.Context) {
	activeSessions, err := sessionService.GetActiveSessions()

	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, err.Error())
	} else {
		c.JSON(http.StatusOK, activeSessions)
	}
}

func (a ApiController) GetSession(c *gin.Context) {
	sessionId := c.Param("sessionId")
	session, err := sessionService.GetSessionById(sessionId)

	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, err.Error())
	} else if session == nil {
		c.AbortWithStatus(http.StatusNotFound)
	} else {
		c.JSON(http.StatusOK, session)
	}
}
