package controller

import (
	"github.com/47-11/spotifete/model"
	"github.com/47-11/spotifete/service"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

type ApiController struct {
	sessionService service.SessionService
}

func (controller ApiController) Index(c *gin.Context) {
	c.String(http.StatusOK, "SpotiFete API v1")
}

func (controller ApiController) GetActiveSessions(c *gin.Context) {
	activeSessions := controller.sessionService.GetActiveSessions()
	c.JSON(http.StatusOK, activeSessions)
}

func (controller ApiController) GetSession(c *gin.Context) {
	sessionId, err := strconv.ParseInt(c.Param("sessionId"), 0, 0)
	session, err := controller.sessionService.GetSessionById(sessionId)

	if err != nil {
		if _, ok := err.(model.EntryNotFoundError); ok {
			c.String(http.StatusNotFound, err.Error())
		} else {
			c.String(http.StatusInternalServerError, err.Error())
		}
	} else {
		c.JSON(http.StatusOK, session)
	}
}
