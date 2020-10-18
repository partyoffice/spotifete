package authentication

import (
	"github.com/47-11/spotifete/database"
	"github.com/47-11/spotifete/database/model"
	. "github.com/47-11/spotifete/error"
	"github.com/jinzhu/gorm"
	"math/rand"
	"time"
)

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

func GetValidSession(sessionId string) *model.LoginSession {
	session := GetSession(sessionId)

	if IsSessionValid(*session) {
		return session
	} else {
		return nil
	}
}

func GetSession(sessionId string) *model.LoginSession {
	var sessions []model.LoginSession
	database.GetConnection().Where(model.LoginSession{SessionId: sessionId}).Find(&sessions)

	if len(sessions) == 0 {
		return nil
	} else {
		return &sessions[0]
	}
}

func NewSession(callbackRedirectUrl string) (newSession model.LoginSession, spotifyAuthUrl string) {
	newSession = model.LoginSession{
		Model:            gorm.Model{},
		SessionId:        newSessionId(),
		UserId:           nil,
		Active:           true,
		CallbackRedirect: callbackRedirectUrl,
	}

	database.GetConnection().Create(&newSession)
	return newSession, authUrlForSession(newSession)
}

func newSessionId() string {
	for {
		b := make([]rune, 256)
		for i := range b {
			b[i] = letterRunes[rand.Intn(len(letterRunes))]
		}
		newSessionId := string(b)

		if !sessionIdExists(newSessionId) {
			return newSessionId
		}
	}
}

func sessionIdExists(sessionId string) bool {
	var count uint

	database.GetConnection().Model(&model.LoginSession{}).Where(model.LoginSession{SessionId: sessionId}).Count(&count)
	return count > 0
}

func isSessionAuthenticatedBySessionId(sessionId string) (isAuthenticated bool, spotifeteError *SpotifeteError) {
	session := GetValidSession(sessionId)
	if session == nil {
		return false, NewUserError("Session not found.")
	}

	return isSessionAuthenticated(*session), nil
}

func isSessionAuthenticated(session model.LoginSession) bool {
	return session.UserId != nil
}

func IsSessionValid(session model.LoginSession) bool {
	return session.Active && session.CreatedAt.AddDate(0, 0, 7).After(time.Now())
}

func InvalidateSession(sessionId string) {
	database.GetConnection().Model(&model.LoginSession{}).Where(model.LoginSession{SessionId: sessionId}).Update("active", false)
}
