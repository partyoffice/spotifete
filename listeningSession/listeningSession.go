package listeningSession

import (
	"errors"
	"fmt"
	"gorm.io/gorm"
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

func RequestSong(session model.FullListeningSession, trackId string, username string) (createdRequest model.SongRequest, spotifeteError *SpotifeteError) {

	requestSongTask := func(tx *gorm.DB) error {
		createdRequest, spotifeteError = createNewSongRequestInTransaction(session, trackId, username, tx)
		if spotifeteError != nil {
			// TODO: improve this when refactoring errors
			return errors.New("rolling back transaction")
		}

		queue, err := GetLimitedQueue(session.SimpleListeningSession, 3)
		if err != nil {
			return err
		}

		go updatePlaylistIfNecessary(session, queue)

		return nil
	}

	err := database.GetConnection().Transaction(requestSongTask)
	if err != nil {
		return model.SongRequest{}, NewInternalError("could not request song", err)
	}

	return createdRequest, nil
}

func createNewSongRequestInTransaction(session model.FullListeningSession, trackId string, username string, tx *gorm.DB) (model.SongRequest, *SpotifeteError) {
	client := Client(session)

	if isTrackInQueue(session.SimpleListeningSession, trackId, tx) {
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

	updatedTrackMetadata, err := AddOrUpdateTrackMetadataInTransaction(*spotifyTrack, tx)
	if err != nil {
		return model.SongRequest{}, NewInternalError("Could not get track metadata", err)
	}

	if spotifyTrack.IsPlayable == nil || !*spotifyTrack.IsPlayable {
		return model.SongRequest{}, NewUserError("Sorry, this track is not available :/")
	}

	weight, err := getRequestCountForUser(session.SimpleListeningSession, "", tx)
	if err != nil {
		return model.SongRequest{}, NewInternalError("Could not get number of requests for user.", err)
	}

	newSongRequest := model.SongRequest{
		BaseModel:      model.BaseModel{},
		SessionId:      session.ID,
		RequestedBy:    username,
		SpotifyTrackId: updatedTrackMetadata.SpotifyTrackId,
		Weight:         weight,
	}

	filter := map[string]interface{}{
		"session_id": session.ID,
		"played":     false,
	}
	queueLength, err := FindSongRequestCountInTransaction(filter, tx)
	if err != nil {
		return model.SongRequest{}, NewInternalError("Could not fetch duplicates from database", err)
	}
	if queueLength < 2 {
		newSongRequest.Locked = true
	}

	err = tx.Create(&newSongRequest).Error
	if err != nil {
		return model.SongRequest{}, NewInternalError("could not save new request", err)
	}

	return newSongRequest, nil
}

func getRequestCountForUser(session model.SimpleListeningSession, requestedBy string, tx *gorm.DB) (int64, error) {

	filter := map[string]interface{}{
		"session_id":   session.ID,
		"requested_by": requestedBy,
	}
	return FindSongRequestCountInTransaction(filter, tx)
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

	if shouldUpdateQueue(session, queue) {
		queue, err = updateQueue(queue)
		if err != nil {
			return NewInternalError("could not update session", err)
		}
	}

	return updatePlaylistIfNecessary(session, queue)
}

func shouldUpdateQueue(session model.FullListeningSession, queue []model.SongRequest) bool {

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

func updateQueue(queue []model.SongRequest) (updatedQueue []model.SongRequest, err error) {

	updateSessionTask := func(tx *gorm.DB) error {
		updatedQueue, err = updateQueueInTransaction(queue, tx)
		return err
	}
	err = database.GetConnection().Transaction(updateSessionTask)

	return updatedQueue, err
}

func updateQueueInTransaction(queue []model.SongRequest, tx *gorm.DB) (updatedQueue []model.SongRequest, err error) {

	err = markPreviousAsPlayed(queue, tx)
	if err != nil {
		return []model.SongRequest{}, err
	}

	err = updateFirst(queue, tx)
	if err != nil {
		return []model.SongRequest{}, err
	}

	err = updateSecondIfPresent(queue, tx)
	if err != nil {
		return []model.SongRequest{}, err
	}

	return queue[1:], nil
}

func markPreviousAsPlayed(queue []model.SongRequest, tx *gorm.DB) error {

	queue[0].Played = true
	return tx.Save(&queue[0]).Error
}

func updateFirst(queue []model.SongRequest, tx *gorm.DB) error {

	queue[1].Locked = true
	queue[1].Weight = 0
	return tx.Save(&queue[1]).Error
}

func updateSecondIfPresent(queue []model.SongRequest, tx *gorm.DB) error {

	if len(queue) >= 3 {
		queue[2].Locked = true
		queue[2].Weight = 1
		return tx.Save(&queue[2]).Error
	} else {
		return nil
	}
}

func updatePlaylistIfNecessary(session model.FullListeningSession, queue []model.SongRequest) *SpotifeteError {

	shouldUpdateSessionPlaylist, spotifeteError := shouldUpdatePlaylist(session, queue)
	if spotifeteError != nil {
		return spotifeteError
	}

	if shouldUpdateSessionPlaylist {
		return updatePlaylist(session, queue)
	}

	return nil
}

func shouldUpdatePlaylist(session model.FullListeningSession, queue []model.SongRequest) (bool, *SpotifeteError) {

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

func updatePlaylist(session model.FullListeningSession, queue []model.SongRequest) *SpotifeteError {

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
