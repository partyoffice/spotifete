package model

type Session struct {
	Uuid   string      `json:"uuid"`
	Active bool        `json:"active"`
	Owner  SpotifyUser `json:"owner"`
}
