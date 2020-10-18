package service

import (
	"github.com/47-11/spotifete/database"
	"github.com/47-11/spotifete/database/model"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"math/rand"
	"sync"
	"time"
)

type loginSessionService struct{}

var loginSessionServiceInstance *loginSessionService
var loginSessionServiceOnce sync.Once

func LoginSessionService() *loginSessionService {
	loginSessionServiceOnce.Do(func() {
		loginSessionServiceInstance = &loginSessionService{}
	})
	return loginSessionServiceInstance
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

func (loginSessionService) sessionIdExists(sessionId string) bool {
	var count uint

	database.GetConnection().Model(&model.LoginSession{}).Where(model.LoginSession{SessionId: sessionId}).Count(&count)
	return count > 0
}

func (s loginSessionService) newSessionId() string {
	for {
		b := make([]rune, 256)
		for i := range b {
			b[i] = letterRunes[rand.Intn(len(letterRunes))]
		}
		newSessionId := string(b)

		if !s.sessionIdExists(newSessionId) {
			return newSessionId
		}
	}
}

func (s loginSessionService) GetSessionBySessionId(sessionId string, requireValid bool) *model.LoginSession {
	var sessions []model.LoginSession
	database.GetConnection().Where(model.LoginSession{SessionId: sessionId}).Find(&sessions)

	if len(sessions) == 1 {
		session := sessions[0]
		if requireValid && !s.IsSessionValid(session) {
			return nil
		} else {
			return &session
		}
	}
	return nil
}

func (s loginSessionService) GetSessionFromCookie(c *gin.Context) *model.LoginSession {
	sessionId, err := c.Cookie("SESSIONID")
	if err != nil || sessionId == "" {
		// No cookie found -> Create new session id and save a new sentry with that id to the database
		return nil
	}

	// Cookie found
	session := s.GetSessionBySessionId(sessionId, true)
	if session != nil {
		// Sesssion found in database
		return session
	} else {
		// Session not found or not valid
		s.InvalidateSession(c)
		return nil
	}
}

func (s loginSessionService) createAndSetNewSession(c *gin.Context) model.LoginSession {
	return s.createAndSetSession(c, s.newSessionId())
}

func (s loginSessionService) createAndSetSession(c *gin.Context, sessionId string) model.LoginSession {
	s.SetSessionCookie(c, sessionId)
	newLoginSession := model.LoginSession{
		Model:     gorm.Model{},
		SessionId: sessionId,
		UserId:    nil,
		Active:    true,
	}
	database.GetConnection().Create(&newLoginSession)

	return newLoginSession
}

func (loginSessionService) SetUserForSession(session model.LoginSession, user model.User) {
	session.UserId = &user.ID
	database.GetConnection().Save(session)
}

func (s loginSessionService) InvalidateSession(c *gin.Context) {
	sessionId, _ := c.Cookie("SESSIONID")
	if sessionId == "" {
		return
	} else {
		c.SetCookie("SESSIONID", "", -1, "/", "", false, true)
		s.InvalidateSessionBySessionId(sessionId)
	}
}

func (loginSessionService) InvalidateSessionBySessionId(sessionId string) {
	database.GetConnection().Model(&model.LoginSession{}).Where(model.LoginSession{SessionId: sessionId}).Update("active", false)
}

func (loginSessionService) IsSessionValid(session model.LoginSession) bool {
	return session.Active && session.CreatedAt.AddDate(0, 0, 7).After(time.Now())
}

func (loginSessionService) SetSessionCookie(c *gin.Context, sessionId string) {
	c.SetCookie("SESSIONID", sessionId, 0, "/", "", false, true)
}
