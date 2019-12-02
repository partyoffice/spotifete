package dto

import "github.com/47-11/spotifete/database/model"

type UserDto struct {
	SpotifyId          string
	SpotifyDisplayName string
}

func (dto UserDto) FromDatabaseModel(databaseModel model.User) UserDto {
	dto.SpotifyId = databaseModel.SpotifyId
	dto.SpotifyDisplayName = databaseModel.SpotifyDisplayName
	return dto
}
