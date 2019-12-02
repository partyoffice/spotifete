package dto

import "github.com/47-11/spotifete/database/model"

type ListeningSessionDto struct {
	Active  bool
	OwnerId uint
	JoinId  string
}

func (self ListeningSessionDto) FromDatabaseModel(databaseModel model.ListeningSession) ListeningSessionDto {
	self.Active = databaseModel.Active
	self.OwnerId = databaseModel.OwnerId
	self.JoinId = *databaseModel.JoinId
	return self
}
