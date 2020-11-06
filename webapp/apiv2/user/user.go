package user

import (
	"fmt"
	"github.com/47-11/spotifete/authentication"
	"github.com/47-11/spotifete/database/model"
	"github.com/47-11/spotifete/shared"
	"github.com/47-11/spotifete/users"
	. "github.com/47-11/spotifete/webapp/apiv2/shared"
	"github.com/gin-gonic/gin"
	"net/http"
)

func getCurrentUser(c *gin.Context) {
	loginSessionId := c.Query("loginSessionId")
	if loginSessionId == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{Message: "Missing query parameter 'loginSessionId'."})
		return
	}

	loginSession := authentication.GetSession(loginSessionId)
	if loginSession == nil || !loginSession.IsValid() {
		c.JSON(http.StatusBadRequest, ErrorResponse{Message: "Invalid login session id."})
		return
	}

	if !loginSession.IsAuthenticated() {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Message: "Login session is not authorized."})
		return
	}

	fullUser := users.FindFullUser(model.SimpleUser{
		BaseModel: model.BaseModel{ID: *loginSession.UserId},
	})
	if fullUser == nil {
		SetJsonError(*shared.NewInternalError(fmt.Sprintf("Could not find full user with ID %d", *loginSession.UserId), nil), c)
		return
	}

	c.JSON(http.StatusOK, fullUser)
}
