package authentication

import (
	"crypto/rand"
	"github.com/47-11/spotifete/database"
	"github.com/47-11/spotifete/database/model"
	. "github.com/47-11/spotifete/shared"
	"math/big"
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

func NewSession(callbackRedirectUrl string) (newSession model.LoginSession, spotifyAuthUrl string, error *SpotifeteError) {
	sessionId, spotifeteError := newSessionId()
	if spotifeteError != nil {
		return model.LoginSession{}, "", spotifeteError
	}

	newSession = model.LoginSession{
		BaseModel:        model.BaseModel{},
		SessionId:        sessionId,
		UserId:           nil,
		Active:           true,
		CallbackRedirect: callbackRedirectUrl,
	}

	database.GetConnection().Create(&newSession)
	return newSession, authUrlForSession(newSession), nil
}

func newSessionId() (string, *SpotifeteError) {
	for {
		newSessionId, spotifeteError := randomSessionId()
		if spotifeteError != nil {
			return "", spotifeteError
		}

		if !sessionIdExists(newSessionId) {
			return newSessionId, nil
		}
	}
}

func randomSessionId() (string, *SpotifeteError) {
	maxRandValue := big.NewInt(int64(len(letterRunes)))

	b := make([]rune, 256)
	for i := range b {
		randInt, err := rand.Int(rand.Reader, maxRandValue)
		if err != nil {
			return "", NewInternalError("Could not generate random int", err)
		}

		b[i] = letterRunes[randInt.Uint64()]
	}

	return string(b), nil
}

func sessionIdExists(sessionId string) bool {
	var count int64

	database.GetConnection().Model(&model.LoginSession{}).Where(model.LoginSession{SessionId: sessionId}).Count(&count)
	return count > 0
}

func InvalidateSession(sessionId string) {
	database.GetConnection().Model(&model.LoginSession{}).Where(model.LoginSession{SessionId: sessionId}).Update("active", false)
}
