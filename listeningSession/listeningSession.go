package listeningSession

import (
	"fmt"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/partyoffice/spotifete/database"
	"github.com/partyoffice/spotifete/database/model"
	. "github.com/partyoffice/spotifete/shared"
	"github.com/partyoffice/spotifete/users"
	"github.com/zmb3/spotify"
)

var numberRunes = []rune("0123456789")

func GetTotalSessionCount() uint {
	var count int64
	database.GetConnection().Model(&model.SimpleListeningSession{}).Count(&count)
	return uint(count)
}

func GetActiveSessionCount() uint {
	var count int64
	database.GetConnection().Model(&model.SimpleListeningSession{}).Where(model.SimpleListeningSession{Active: true}).Count(&count)
	return uint(count)
}

func FindSimpleListeningSession(filter model.SimpleListeningSession) *model.SimpleListeningSession {
	listeningSessions := FindSimpleListeningSessions(filter)

	resultCount := len(listeningSessions)
	if resultCount == 1 {
		return &listeningSessions[0]
	} else if resultCount == 0 {
		return nil
	} else {
		NewInternalError(fmt.Sprintf("Got more than one result for filter %v", filter), nil)
		return nil
	}
}

func FindSimpleListeningSessions(filter model.SimpleListeningSession) []model.SimpleListeningSession {
	var listeningSessions []model.SimpleListeningSession
	database.GetConnection().Where(filter).Find(&listeningSessions)
	return listeningSessions
}

func FindFullListeningSession(filter model.SimpleListeningSession) *model.FullListeningSession {
	listeningSessions := FindFullListeningSessions(filter)

	resultCount := len(listeningSessions)
	if resultCount == 1 {
		return &listeningSessions[0]
	} else if resultCount == 0 {
		return nil
	} else {
		NewInternalError(fmt.Sprintf("Got more than one result for filter %v", filter), nil)
		return nil
	}
}

func FindFullListeningSessions(filter model.SimpleListeningSession) []model.FullListeningSession {
	var listeningSessions []model.FullListeningSession
	database.GetConnection().Where(filter).Joins("Owner").Joins("FallbackPlaylistMetadata").Find(&listeningSessions)
	return listeningSessions
}

func NewSession(user model.SimpleUser, title string) (*model.SimpleListeningSession, *SpotifeteError) {

	cleanedTitle, spotifeteError := cleanTitle(title)
	if spotifeteError != nil {
		return nil, spotifeteError
	}

	joinId := newJoinId()
	queuePlaylist, spotifeteError := createPlaylistForSession(joinId, cleanedTitle, user)

	listeningSession := model.SimpleListeningSession{
		BaseModel:               model.BaseModel{},
		Active:                  true,
		OwnerId:                 user.ID,
		JoinId:                  joinId,
		QueuePlaylistId:         queuePlaylist.ID.String(),
		Title:                   title,
		FallbackPlaylistShuffle: true,
	}

	database.GetConnection().Create(&listeningSession)

	return &listeningSession, nil
}

func cleanTitle(rawTitle string) (cleanedTitle string, err *SpotifeteError) {

	trimmedTitle := strings.TrimSpace(rawTitle)
	if len(trimmedTitle) == 0 {
		return "", NewUserError("Session title must not be empty.")
	}

	if len(trimmedTitle) > 100 {
		return "", NewUserError("Session title must not be longer than 100 characters.")
	}

	return trimmedTitle, nil
}

func newJoinId() string {
	rand.Seed(time.Now().UnixNano())

	for {
		b := make([]rune, 8)
		for i := range b {
			b[i] = numberRunes[rand.Intn(len(numberRunes))]
		}
		newJoinId := string(b)

		if joinIdFree(newJoinId) {
			return newJoinId
		}
	}
}

func joinIdFree(joinId string) bool {
	existingListeningSession := FindSimpleListeningSession(model.SimpleListeningSession{
		JoinId: joinId,
		Active: true,
	})

	return existingListeningSession == nil
}

func CloseSession(user model.SimpleUser, joinId string) *SpotifeteError {
	session := FindSimpleListeningSession(model.SimpleListeningSession{
		JoinId: joinId,
		Active: true,
	})
	if session == nil {
		return NewUserError("Unknown listening session.")
	}

	if user.ID != session.OwnerId {
		return NewUserError("Only the session owner can close a session.")
	}

	session.Active = false
	database.GetConnection().Save(&session)
	// TODO: Use a transaction here

	client := users.Client(user)
	// TODO: Only try to unfollow playlist if owner is still following it.
	err := client.UnfollowPlaylist(spotify.ID(user.SpotifyId), spotify.ID(session.QueuePlaylistId))
	if err != nil {
		return NewError("Could not unfollow (delete) playlist.", err, http.StatusInternalServerError)
	}

	// Create rewind playlist if any tracks were requested
	distinctRequestedTracks := GetDistinctRequestedTracks(*session)
	if len(distinctRequestedTracks) > 0 {
		rewindPlaylist, err := client.CreatePlaylistForUser(user.SpotifyId, fmt.Sprintf("%s Rewind - SpotiFete", session.Title), fmt.Sprintf("Rewind playlist for your session %s. This contains all the songs that were requested.", session.Title), false)
		if err != nil {
			return NewError("Could not create rewind playlist.", err, http.StatusInternalServerError)
		}

		go func() {
			var page []spotify.ID
			for _, track := range distinctRequestedTracks {
				page = append(page, track)

				if len(page) == 100 {
					_, err = client.AddTracksToPlaylist(rewindPlaylist.ID, page...)
					if err != nil {
						_ = NewInternalError(fmt.Sprintf("Could not add tracks to rewind playlist. Session: %d | Rewind playlist: %s | Tracks: %s", session.ID, rewindPlaylist.ID, page), err)
					}
					page = []spotify.ID{}
				}
			}

			if len(page) > 0 {
				_, err = client.AddTracksToPlaylist(rewindPlaylist.ID, page...)
				if err != nil {
					_ = NewInternalError(fmt.Sprintf("Could not add tracks to rewind playlist. Session: %d | Rewind playlist: %s | Tracks: %s", session.ID, rewindPlaylist.ID, page), err)
				}
			}
		}()
	}

	return nil
}

// TODO: This is probably not thread-safe
func RequestSong(session model.FullListeningSession, trackId string, username string) (model.SongRequest, *SpotifeteError) {
	client := Client(session)

	// Prevent duplicates
	if IsTrackInQueue(session.SimpleListeningSession, trackId) {
		return model.SongRequest{}, NewUserError("This tack is already in the queue.")
	}

	currentUser, err := client.CurrentUser()
	if err != nil {
		return model.SongRequest{}, NewError("Could not get user information on session owner from Spotify.", err, http.StatusInternalServerError)
	}

	spotifyTrack, err := client.GetTrackOpt(spotify.ID(trackId), &spotify.Options{
		Country: &currentUser.Country,
	})
	if err != nil {
		return model.SongRequest{}, NewError("Could not get track information from spotify.", err, http.StatusInternalServerError)
	}
	updatedTrackMetadata := AddOrUpdateTrackMetadata(*spotifyTrack)

	if spotifyTrack.IsPlayable == nil || !*spotifyTrack.IsPlayable {
		return model.SongRequest{}, NewUserError("Sorry, this track is not available :/")
	}

	queue, err := GetLimitedQueue(session.SimpleListeningSession, 2)
	if err != nil {
		return model.SongRequest{}, NewInternalError("Could not fetch queue from database", err)
	}

	locked := len(queue) < 2

	weight, err := getRequestCountForUser(session.SimpleListeningSession, "")
	if err != nil {
		return model.SongRequest{}, NewInternalError("Could not get number of requests for user.", err)
	}

	newSongRequest := model.SongRequest{
		BaseModel:      model.BaseModel{},
		SessionId:      session.ID,
		RequestedBy:    username,
		SpotifyTrackId: updatedTrackMetadata.SpotifyTrackId,
		Played:         false,
		Locked:         locked,
		Weight:         weight,
	}

	database.GetConnection().Create(&newSongRequest)

	return newSongRequest, updateSessionPlaylistIfNecessary(session)
}

func getRequestCountForUser(session model.SimpleListeningSession, requestedBy string) (int64, error) {

	filter := map[string]interface{}{
		"session_id":   session.ID,
		"requested_by": requestedBy,
	}
	return FindSongRequestCount(filter)
}

func UpdateSessionIfNecessary(session model.FullListeningSession) *SpotifeteError {

	queue, err := GetLimitedQueue(session.SimpleListeningSession, 3)
	if err != nil {
		return NewInternalError("Could not fetch queue from database", err)
	}

	queue, spotifeteError := addFallbackTrackIfNecessary(session, queue)
	if spotifeteError != nil {
		return spotifeteError
	}

	if isSessionUpdateNecessary(session, queue) {
		return updateSession(queue)
	}

	return updateSessionPlaylistIfNecessary(session)
}

func isSessionUpdateNecessary(session model.FullListeningSession, queue []model.SongRequest) bool {

	if len(queue) < 2 {
		return false
	}

	currentlyPlaying, err := Client(session).PlayerCurrentlyPlaying()
	if err != nil {
		NewInternalError("Could not get currently playing track from Spotify.", err)
		currentlyPlaying = nil
	}

	if currentlyPlaying == nil || currentlyPlaying.Item == nil {
		return false
	}

	nextRequest := queue[1]
	return isSessionPlaying(session, currentlyPlaying.PlaybackContext) && nextRequest.SpotifyTrackId == currentlyPlaying.Item.ID.String()
}

func isSessionPlaying(session model.FullListeningSession, playbackContext spotify.PlaybackContext) bool {
	return playbackContext.Type == "playlist" && strings.HasSuffix(string(playbackContext.URI), session.QueuePlaylistId)
}

func updateSession(queue []model.SongRequest) *SpotifeteError {

	queue[0].Played = true
	err := database.GetConnection().Save(&queue[0]).Error
	if err != nil {
		return NewInternalError("could not save song request after marking as played", err)
	}

	if len(queue) >= 3 {
		queue[2].Locked = true
		err = database.GetConnection().Save(&queue[2]).Error
		if err != nil {
			return NewInternalError("could not save song request after marking as locked", err)
		}
	}

	return nil
}

func updateSessionPlaylistIfNecessary(session model.FullListeningSession) *SpotifeteError {

	queue, err := GetLimitedQueue(session.SimpleListeningSession, 2)
	if err != nil {
		return NewInternalError("could not fetch limited queue from database", err)
	}

	shouldUpdateSessionPlaylist, spotifeteError := shouldUpdateSessionPlaylist(session, queue)
	if spotifeteError != nil {
		return spotifeteError
	}

	if shouldUpdateSessionPlaylist {
		return updateSessionPlaylist(session, queue)
	}

	return nil
}

func shouldUpdateSessionPlaylist(session model.FullListeningSession, queue []model.SongRequest) (bool, *SpotifeteError) {

	queueLength := len(queue)
	if queueLength == 0 {
		return false, nil
	}

	client := Client(session)

	playlist, err := client.GetPlaylist(spotify.ID(session.QueuePlaylistId))
	if err != nil {
		return false, NewInternalError("Could not get playlist information from Spotify.", err)
	}

	playlistTracks := playlist.Tracks.Tracks
	playlistLength := len(playlistTracks)

	if playlistLength <= 0 {
		return true, nil
	}

	if queue[0].SpotifyTrackId != playlistTracks[0].Track.ID.String() {
		return true, nil
	}

	if queueLength > 1 {
		if playlistLength <= 1 {
			return true, nil
		}

		if queue[1].SpotifyTrackId != playlistTracks[1].Track.ID.String() {
			return true, nil
		}
	}

	return false, nil
}

func updateSessionPlaylist(session model.FullListeningSession, queue []model.SongRequest) *SpotifeteError {

	client := Client(session)
	playlistId := spotify.ID(session.QueuePlaylistId)
	firstTrackId := spotify.ID(queue[0].SpotifyTrackId)

	var err error
	if len(queue) > 1 {
		secondTrackId := spotify.ID(queue[1].SpotifyTrackId)
		err = client.ReplacePlaylistTracks(playlistId, firstTrackId, secondTrackId)
	} else {
		err = client.ReplacePlaylistTracks(playlistId, firstTrackId)
	}

	if err != nil {
		return NewError("Could not update tracks in playlist.", err, http.StatusInternalServerError)
	}

	return nil
}

func NewQueuePlaylist(session model.FullListeningSession) *SpotifeteError {

	owner := session.Owner
	client := users.Client(owner)

	err := client.UnfollowPlaylist(spotify.ID(owner.SpotifyId), spotify.ID(session.QueuePlaylistId))
	if err != nil {
		return NewError("Could not unfollow old playlist.", err, http.StatusInternalServerError)
	}

	newPlaylist, spotifeteError := createPlaylistForSession(session.JoinId, session.Title, owner)
	if spotifeteError != nil {
		return spotifeteError
	}

	session.QueuePlaylistId = newPlaylist.ID.String()
	database.GetConnection().Save(&session)

	return nil
}

func RefollowQueuePlaylist(session model.FullListeningSession) *SpotifeteError {

	owner := session.Owner
	client := users.Client(owner)

	ownerId := spotify.ID(owner.SpotifyId)
	queuePlaylistId := spotify.ID(session.QueuePlaylistId)

	err := client.UnfollowPlaylist(ownerId, queuePlaylistId)
	if err != nil {
		return NewError("Could not unfollow playlist.", err, http.StatusInternalServerError)
	}

	time.Sleep(1 * time.Second)

	err = client.FollowPlaylist(ownerId, queuePlaylistId, false)
	if err == nil {
		return nil
	} else {
		return NewError("Could not unfollow playlist.", err, http.StatusInternalServerError)
	}
}
