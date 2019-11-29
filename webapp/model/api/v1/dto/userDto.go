package dto

import "github.com/47-11/spotifete/database/model"

type UserDto struct {
	SpotifyId          string
	SpotifyDisplayName string
}

func (self UserDto) FromDatabaseModel(databaseModel model.User) UserDto {
	self.SpotifyId = databaseModel.SpotifyId
	self.SpotifyDisplayName = databaseModel.SpotifyDisplayName
	return self
}
