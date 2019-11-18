package service

import (
	"github.com/47-11/spotifete/database/model"
	"github.com/gin-gonic/gin"
	"math/rand"
	"sync"
)

type loginSessionService struct {
	userService   UserService
	loginSessions map[string]*uint
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
var instance *loginSessionService
var once sync.Once

func LoginSessionService() *loginSessionService {
	once.Do(func() {
		instance = &loginSessionService{
			loginSessions: make(map[string]*uint),
		}
	})
	return instance
}

func (s loginSessionService) GetUserForLoginSession(sessionId string) (*model.User, error) {
	if userId, found := s.loginSessions[sessionId]; found {
		return s.userService.GetUserById(*userId)
	} else {
		return nil, nil
	}
}

func (s loginSessionService) sessionIdExists(sessionId string) bool {
	_, exists := s.loginSessions[sessionId]
	return exists
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

func (s loginSessionService) GetOrCreateSessionId(c *gin.Context, newUserId *uint) (*string, *uint) {
	sessionId, err := c.Cookie("SESSIONID")
	if err != nil {
		// No cookie found -> Set session
		newSessionId := s.newSessionId()
		c.SetCookie("SESSIONID", newSessionId, 0, "/", "", false, true)
		s.loginSessions[newSessionId] = newUserId

		return &newSessionId, newUserId
	} else {
		// Cookie found
		if sessionUserId, found := s.loginSessions[sessionId]; found && sessionUserId != nil {
			return &sessionId, sessionUserId
		} else {
			s.loginSessions[sessionId] = newUserId
			return &sessionId, newUserId
		}
	}
}

func (s loginSessionService) DeleteSessionId(c *gin.Context) {
	sessionId, err := c.Cookie("SESSIONID")
	if err == nil {
		delete(s.loginSessions, sessionId)
	}
}
