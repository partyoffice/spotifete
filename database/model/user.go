package model

import (
	"golang.org/x/oauth2"
	"time"
)

type SimpleUser struct {
	BaseModel
	SpotifyId           string
	SpotifyDisplayName  string
	SpotifyAccessToken  string    `json:"-"`
	SpotifyRefreshToken string    `json:"-"`
	SpotifyTokenType    string    `json:"-"`
	SpotifyTokenExpiry  time.Time `json:"-"`
}

func (SimpleUser) TableName() string {
	return "users"
}

func (u SimpleUser) GetToken() *oauth2.Token {
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

func (u SimpleUser) SetToken(token *oauth2.Token) SimpleUser {
	u.SpotifyAccessToken = token.AccessToken
	u.SpotifyRefreshToken = token.RefreshToken
	u.SpotifyTokenType = token.TokenType
	u.SpotifyTokenExpiry = token.Expiry

	return u
}

type FullUser struct {
	SimpleUser

	ListeningSessions []SimpleListeningSession `gorm:"foreignKey:owner_id"`
}

func (FullUser) TableName() string {
	return "users"
}
