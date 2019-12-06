package dto

import (
	"github.com/47-11/spotifete/model/database"
)

type UserDto struct {
	SpotifyId          string                `json:"spotifyId"`
	SpotifyDisplayName string                `json:"spotifyDisplayName"`
	ListeningSessions  []ListeningSessionDto `json:"listeningSessions"`
}

func (dto UserDto) FromDatabaseModel(databaseModel *database.User) UserDto {
	dto.SpotifyId = databaseModel.SpotifyId
	dto.SpotifyDisplayName = databaseModel.SpotifyDisplayName
	dto.ListeningSessions = []ListeningSessionDto{}
	return dto
}
