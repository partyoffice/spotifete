package service

import (
	"errors"
	"github.com/47-11/spotifete/database"
	"github.com/47-11/spotifete/database/model"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"math/rand"
	"sync"
	"time"
)

type loginSessionService struct {
	userService UserService
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
var instance *loginSessionService
var once sync.Once

func LoginSessionService() *loginSessionService {
	once.Do(func() {
		instance = &loginSessionService{
			userService: UserService{},
		}
	})
	return instance
}

func (s loginSessionService) GetUserForLoginSession(sessionId string) (*model.User, error) {
	var sessions []model.LoginSession
	database.Connection.Where("session_id = ?", sessionId).Find(&sessions)
	if len(sessions) == 1 {
		return s.userService.GetUserById(*sessions[0].UserId)
	} else {
		return nil, nil
	}
}

func (s loginSessionService) sessionIdExists(sessionId string) bool {
	var count uint
	database.Connection.Model(&model.LoginSession{}).Where("session_id = ?", sessionId).Count(&count)
	return count == 1
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

func (s loginSessionService) GetSessionBySessionId(sessionId string) *model.LoginSession {
	sessions := []model.LoginSession{}
	database.Connection.Where("session_id = ?", sessionId).Find(&sessions)

	if len(sessions) == 1 {
		return &sessions[0]
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
	session := s.GetSessionBySessionId(sessionId)
	if session != nil {
		// Sesssion found in database
		if s.IsSessionValid(*session) {
			return session
		} else {
			return nil
		}

	} else {
		// The session id from the cookie could not be found in database -> this normally should not happen and
		// could be an indicator for a malicious attack. For now just return nil
		// TODO: Do something smart when this happens
		return nil
	}
}

func (s loginSessionService) createAndSetNewSession(c *gin.Context) model.LoginSession {
	return s.createAndSetNewession(c, s.newSessionId())
}

func (s loginSessionService) createAndSetNewession(c *gin.Context, sessionId string) model.LoginSession {
	s.SetSessionCookie(c, sessionId)
	newLoginSession := model.LoginSession{
		Model:     gorm.Model{},
		SessionId: sessionId,
		UserId:    nil,
		Active:    true,
	}
	database.Connection.Create(&newLoginSession)

	return newLoginSession
}

func (s loginSessionService) SetUserForSession(session model.LoginSession, user model.User) {
	// TODO: Invalidate all other sessions for that user
	session.UserId = &user.ID
	database.Connection.Save(session)
}

func (s loginSessionService) InvalidateSession(c *gin.Context) error {
	sessionId, err := c.Cookie("SESSIONID")
	if err == nil {
		c.SetCookie("SESSIONID", "", -1, "/", "", false, true)
		return s.InvalidateSessionBySessionId(sessionId)
	}

	return errors.New("session cookie not present")
}

func (s loginSessionService) InvalidateSessionBySessionId(sessionId string) error {
	rowsAffected := database.Connection.Model(&model.LoginSession{}).Where("session_id = ?", sessionId).Update("active", false).RowsAffected
	if rowsAffected > 0 {
		return nil
	} else {
		return errors.New("session id not found")
	}
}

func (s loginSessionService) IsSessionValid(session model.LoginSession) bool {
	return session.Active && session.CreatedAt.AddDate(0, 1, 0).After(time.Now())
}

func (s loginSessionService) SetSessionCookie(c *gin.Context, sessionId string) {
	c.SetCookie("SESSIONID", sessionId, 0, "/", "", false, true)
}
