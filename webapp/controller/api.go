package controller

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
)

type ApiController struct{}

func (a ApiController) Index(c *gin.Context) {
	c.String(http.StatusOK, "SpotiFete API v1")
}

func (a ApiController) GetSessions(c *gin.Context) {
	fmt.Println("Requested all sessions")
}

func (a ApiController) GetSession(c *gin.Context) {
	sessionId := c.Param("sessionId")
	fmt.Printf("Reqested session with id %s", sessionId)
}
