package model

import (
	"golang.org/x/oauth2"
	"gorm.io/gorm"
	"time"
)

type User struct {
	gorm.Model
	SpotifyId           string
	SpotifyDisplayName  string
	SpotifyAccessToken  string
	SpotifyRefreshToken string
	SpotifyTokenType    string
	SpotifyTokenExpiry  time.Time

	ListeningSessions *[]ListeningSession `gorm:"foreignKey:owner_id"`
}

func (u User) GetToken() *oauth2.Token {
	if len(u.SpotifyAccessToken) > 0 && len(u.SpotifyRefreshToken) > 0 && len(u.SpotifyTokenType) > 0 {
		return &oauth2.Token{
			AccessToken:  u.SpotifyAccessToken,
			TokenType:    u.SpotifyTokenType,
			RefreshToken: u.SpotifyRefreshToken,
			Expiry:       u.SpotifyTokenExpiry,
		}
	} else {
		return nil
	}
}
