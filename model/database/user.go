package database

import (
	"github.com/jinzhu/gorm"
	"golang.org/x/oauth2"
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
}

func (u User) GetToken() *oauth2.Token {
	if len(u.SpotifyAccessToken) > 0 && len(u.SpotifyRefreshToken) > 0 && len(u.SpotifyTokenType) > 0 {
		return &oauth2.Token{
			AccessToken:  u.SpotifyAccessToken,
			TokenType:    u.SpotifyTokenType,
			RefreshToken: u.SpotifyRefreshToken,
			// TODO: Find a proper fix instead of this horrible hack
			Expiry: u.SpotifyTokenExpiry.Add(-1 * time.Hour).UTC(),
		}
	} else {
		return nil
	}
}
