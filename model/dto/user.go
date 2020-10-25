package dto

import "github.com/47-11/spotifete/database/model"

type SimpleUserDto struct {
	SpotifyId          string `json:"spotifyId"`
	SpotifyDisplayName string `json:"spotifyDisplayName"`
}

func NewSimpleUserDto(simpleUser model.SimpleUser) SimpleUserDto {
	return SimpleUserDto{
		SpotifyId:          simpleUser.SpotifyId,
		SpotifyDisplayName: simpleUser.SpotifyDisplayName,
	}
}

type FullUserDto struct {
	SimpleUserDto
	ListeningSessions []ListeningSessionDto `json:"listeningSessions,omitempty"`
}

func NewFullUserDto(fullUser model.FullUser) FullUserDto {
	return FullUserDto{
		SimpleUserDto:     NewSimpleUserDto(fullUser.SimpleUser),
		ListeningSessions: []ListeningSessionDto{},
	}
}
