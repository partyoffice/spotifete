package controller

import (
	"github.com/47-11/spotifete/service"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

type TemplateController struct {
	sessionService service.SessionService
}

func (controller TemplateController) Index(c *gin.Context) {
	c.HTML(http.StatusOK, "index.html", gin.H{
		"time":               time.Now(),
		"activeSessionCount": controller.sessionService.GetActiveSessionCount(),
		"totalSessionCount":  controller.sessionService.GetTotalSessionCount(),
	})
}
