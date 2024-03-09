package authentication

import (
	"github.com/gin-gonic/gin"
	"github.com/partyoffice/spotifete/database/model"
)

const sessionCookieName = "SF_SESSION_ID"

func GetValidSessionFromCookie(c *gin.Context) *model.LoginSession {
	sessionId := GetSessionIdFromCookie(c)
	if sessionId == nil {
		return nil
	}

	session := GetSession(*sessionId)
	if session != nil && session.IsValid() {
		return session
	} else {
		InvalidateSession(*sessionId)
		RemoveCookie(c)
		return nil
	}
}

func GetSessionIdFromCookie(c *gin.Context) *string {
	sessionId, err := c.Cookie(sessionCookieName)
	if err != nil || sessionId == "" {
		return nil
	}

	return &sessionId
}

func SetCookie(c *gin.Context, sessionId string) {
	c.SetCookie(sessionCookieName, sessionId, 0, "/", "", false, true)
}

func RemoveCookie(c *gin.Context) {
	c.SetCookie(sessionCookieName, "", -1, "/", "", false, true)
}
