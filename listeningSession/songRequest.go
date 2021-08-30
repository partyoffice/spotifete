package listeningSession

import (
	"fmt"
	"sort"
	"time"

	"github.com/partyoffice/spotifete/database"
	"github.com/partyoffice/spotifete/database/model"
	. "github.com/partyoffice/spotifete/shared"
	"github.com/zmb3/spotify"
)

func FindSongRequest(filter model.SongRequest) *model.SongRequest {
	songRequests := FindSongRequests(filter)

	resultCount := len(songRequests)
	if resultCount == 1 {
		return &songRequests[0]
	} else if resultCount == 0 {
		return nil
	} else {
		NewInternalError(fmt.Sprintf("Got more than one result for filter %v", filter), nil)
		return nil
	}
}

func FindSongRequests(filter model.SongRequest) []model.SongRequest {
	var songRequests []model.SongRequest
	database.GetConnection().Where(filter).Joins("TrackMetadata").Find(&songRequests)
	return songRequests
}

func GetCurrentlyPlayingRequest(session model.SimpleListeningSession) *model.SongRequest {
	return FindSongRequest(model.SongRequest{
		SessionId: session.ID,
		Status:    model.StatusCurrentlyPlaying,
	})
}

func GetUpNextRequest(session model.SimpleListeningSession) *model.SongRequest {
	return FindSongRequest(model.SongRequest{
		SessionId: session.ID,
		Status:    model.StatusUpNext,
	})
}

func GetSessionQueueInDemocraticOrder(session model.SimpleListeningSession) []model.SongRequest {
	queue := FindSongRequests(model.SongRequest{
		SessionId: session.ID,
		Status:    model.StatusInQueue,
	})

	sort.SliceStable(queue, func(i, j int) bool {
		return queue[i].ID < queue[j].ID
	})

	return queue
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
	database.GetConnection().Where("status != 'PLAYED' AND session_id = ? AND spotify_track_id = ?", session.ID, trackId).Find(&duplicateRequestsForTrack)
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
	requestToDelete := FindSongRequest(model.SongRequest{
		SessionId:      session.ID,
		SpotifyTrackId: spotifyTrackId,
	})
	if requestToDelete == nil {
		return NewUserError("Request not found in queue.")
	}

	if requestToDelete.Status != model.StatusInQueue {
		return NewUserError("The request must be in the queue to be deleted.")
	}

	database.GetConnection().Delete(requestToDelete)

	return nil
}
