package listeningSession

import (
	"fmt"
	"gorm.io/gorm"
	"time"

	"github.com/partyoffice/spotifete/database"
	"github.com/partyoffice/spotifete/database/model"
	. "github.com/partyoffice/spotifete/shared"
	"github.com/zmb3/spotify"
)

func FindSongRequest(filter interface{}) (*model.SongRequest, error) {

	query := database.GetConnection().Where(filter)
	songRequests, err := FindSongRequests(query)
	if err != nil {
		return nil, err
	}

	resultCount := len(songRequests)
	if resultCount == 1 {
		return &songRequests[0], nil
	} else if resultCount == 0 {
		return nil, nil
	} else {
		return nil, fmt.Errorf("got more than one result")
	}
}

func FindSongRequests(query *gorm.DB) ([]model.SongRequest, error) {

	var songRequests []model.SongRequest
	err := query.Joins("TrackMetadata").Find(&songRequests).Error
	return songRequests, err
}

func FindSongRequestCount(filter interface{}) (int64, error) {

	var count int64
	err := database.GetConnection().Model(model.SongRequest{}).Where(filter).Count(&count).Error
	return count, err
}

func GetFullQueue(session model.SimpleListeningSession) ([]model.SongRequest, error) {

	query := buildGetQueueQuery(session)
	return FindSongRequests(query)
}

func GetLimitedQueue(session model.SimpleListeningSession, limit int) ([]model.SongRequest, error) {

	query := buildGetQueueQuery(session).Limit(limit)
	return FindSongRequests(query)
}

func buildGetQueueQuery(session model.SimpleListeningSession) *gorm.DB {

	filter := map[string]interface{}{
		"session_id": session.ID,
		"played":     false,
	}

	return database.GetConnection().Where(filter).Order("locked desc, weight asc, created_at asc")
}

func GetQueueLastUpdated(session model.SimpleListeningSession) time.Time {
	var requests []model.SongRequest
	database.GetConnection().Where(model.SongRequest{
		SessionId: session.ID,
	}).Order("updated_at desc").Find(&requests)

	if len(requests) == 0 {
		// No requests found -> Use creation of session
		return session.UpdatedAt
	} else {
		return requests[0].UpdatedAt
	}
}

func IsTrackInQueue(session model.SimpleListeningSession, trackId string) bool {
	var duplicateRequestsForTrack []model.SongRequest
	database.GetConnection().Where("played = false AND session_id = ? AND spotify_track_id = ?", session.ID, trackId).Find(&duplicateRequestsForTrack)
	return len(duplicateRequestsForTrack) > 0
}

func GetDistinctRequestedTracks(session model.SimpleListeningSession) (trackIds []spotify.ID) {
	type Result struct {
		SpotifyTrackId string
	}

	var results []Result
	database.GetConnection().Table("song_requests").Select("distinct spotify_track_id").Where(model.SongRequest{
		SessionId: session.ID,
	}).Scan(&results)

	for _, result := range results {
		trackIds = append(trackIds, spotify.ID(result.SpotifyTrackId))
	}

	return
}

func DeleteRequestFromQueue(session model.SimpleListeningSession, spotifyTrackId string) *SpotifeteError {

	filter := map[string]interface{}{
		"session_id":       session.ID,
		"spotify_track_id": spotifyTrackId,
		"played":           false,
	}

	requestToDelete, err := FindSongRequest(filter)
	if err != nil {
		return NewInternalError("Could not fetch request to delete", err)
	}
	if requestToDelete == nil {
		return NewUserError("Request not found in queue.")
	}
	if requestToDelete.Locked {
		return NewUserError("This request cannot be deleted.")
	}

	database.GetConnection().Delete(requestToDelete)

	return nil
}
