package model

type Session struct {
	Uuid   string
	Active bool
	Owner  SpotifyUser
}
