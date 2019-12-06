package dto

import "github.com/47-11/spotifete/database/model"

type ListeningSessionDto struct {
	Active  bool   `json:"active"`
	OwnerId uint   `json:"ownerId"`
	JoinId  string `json:"joinId"`
}

func (dto ListeningSessionDto) FromDatabaseModel(databaseModel model.ListeningSession) ListeningSessionDto {
	dto.Active = databaseModel.Active
	dto.OwnerId = databaseModel.OwnerId
	dto.JoinId = *databaseModel.JoinId
	return dto
}
