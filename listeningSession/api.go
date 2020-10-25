package listeningSession

import (
	"github.com/47-11/spotifete/database/model"
	"github.com/47-11/spotifete/shared"
	"github.com/gin-gonic/gin"
	"net/http"
)

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
