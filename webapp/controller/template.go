package controller

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

type TemplateController struct{}

func (a TemplateController) Index(c *gin.Context) {
	c.HTML(http.StatusOK, "index.html", gin.H{
		"time": time.Now(),
	})
}
