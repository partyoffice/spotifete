package dto

type UserDto struct {
	SpotifyId          string                `json:"spotifyId"`
	SpotifyDisplayName string                `json:"spotifyDisplayName"`
	ListeningSessions  []ListeningSessionDto `json:"listeningSessions"`
}
