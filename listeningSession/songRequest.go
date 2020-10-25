package listeningSession

import (
	"github.com/47-11/spotifete/database"
	"github.com/47-11/spotifete/database/model"
)

func FindSongRequest(filter model.SongRequest) *model.SongRequest {
	songRequests := FindSongRequests(filter)

	if len(songRequests) == 1 {
		return &songRequests[0]
	} else {
		return nil
	}
}

func FindSongRequests(filter model.SongRequest) []model.SongRequest {
	var songRequests []model.SongRequest
	database.GetConnection().Where(filter).Find(&songRequests)
	return songRequests
}
