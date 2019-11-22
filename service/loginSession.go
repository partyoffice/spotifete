package service

import (
	"github.com/47-11/spotifete/database"
	"github.com/47-11/spotifete/database/model"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"math/rand"
	"sync"
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

func (s loginSessionService) GetOrCreateSessionId(c *gin.Context, newUserId *uint) model.LoginSession {
	sessionId, err := c.Cookie("SESSIONID")
	if err != nil {
		// No cookie found -> Set session
		newSessionId := s.newSessionId()
		c.SetCookie("SESSIONID", newSessionId, 0, "/", "", false, true)
		newLoginSession := model.LoginSession{
			Model:     gorm.Model{},
			SessionId: newSessionId,
			UserId:    newUserId,
			Active:    true,
		}
		database.Connection.Create(&newLoginSession)

		return newLoginSession
	} else {
		// Cookie found
		var sessions []model.LoginSession
		database.Connection.Where("session_id = ?", sessionId).Find(&sessions)

		if len(sessions) == 1 {
			// Sesssion found in database -> Just return it
			return sessions[0]
		} else {
			// Session not found in database -> this normally should not happen, but create it with an enpty user nonetheless
			newLoginSession := model.LoginSession{
				Model:     gorm.Model{},
				SessionId: sessionId,
				UserId:    nil,
				Active:    true,
			}
			database.Connection.Create(&newLoginSession)
			return newLoginSession
		}
	}
}

func (s loginSessionService) InvalidateSession(c *gin.Context) {
	sessionId, err := c.Cookie("SESSIONID")
	if err == nil {
		database.Connection.Model(&model.LoginSession{}).Where("session_id = ?", sessionId).Update("active", false)
	}
}
