package authentication

import (
	"github.com/47-11/spotifete/database"
	"github.com/47-11/spotifete/database/model"
	"math/rand"
	"time"
)

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

func GetSession(sessionId string) *model.LoginSession {
	var sessions []model.LoginSession
	database.GetConnection().Where("session_id = ?", sessionId).Joins("User").Find(&sessions)

	if len(sessions) == 0 {
		return nil
	} else {
		return &sessions[0]
	}
}

func NewSession(callbackRedirectUrl string) (newSession model.LoginSession, spotifyAuthUrl string) {
	newSession = model.LoginSession{
		BaseModel:        model.BaseModel{},
		SessionId:        newSessionId(),
		UserId:           nil,
		Active:           true,
		CallbackRedirect: callbackRedirectUrl,
	}

	database.GetConnection().Create(&newSession)
	return newSession, authUrlForSession(newSession)
}

func newSessionId() string {
	rand.Seed(time.Now().UnixNano())

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
	var count int64

	database.GetConnection().Model(&model.LoginSession{}).Where(model.LoginSession{SessionId: sessionId}).Count(&count)
	return count > 0
}

func InvalidateSession(sessionId string) {
	database.GetConnection().Model(&model.LoginSession{}).Where(model.LoginSession{SessionId: sessionId}).Update("active", false)
}
