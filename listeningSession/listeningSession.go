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
func RequestSong(session model.FullListeningSession, trackId string) (model.SongRequest, *SpotifeteError) {
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

	// Check if we have to add the request to the queue or play it immediately
	currentlyPlayingRequest := GetCurrentlyPlayingRequest(session.SimpleListeningSession)
	upNextRequest := GetUpNextRequest(session.SimpleListeningSession)

	var newRequestStatus model.SongRequestStatus
	if currentlyPlayingRequest == nil {
		newRequestStatus = model.StatusCurrentlyPlaying
	} else if upNextRequest == nil {
		newRequestStatus = model.StatusUpNext
	} else {
		newRequestStatus = model.StatusInQueue
	}

	newSongRequest := model.SongRequest{
		BaseModel:      model.BaseModel{},
		SessionId:      session.ID,
		UserId:         nil,
		SpotifyTrackId: updatedTrackMetadata.SpotifyTrackId,
		Status:         newRequestStatus,
	}

	database.GetConnection().Create(&newSongRequest)

	return newSongRequest, updateSessionPlaylistIfNecessary(session)
}

func UpdateSessionIfNecessary(session model.FullListeningSession) *SpotifeteError {
	upNextRequest := GetUpNextRequest(session.SimpleListeningSession)

	trackAdded, spotifeteError := addFallbackTrackIfNecessary(session, upNextRequest)
	if spotifeteError != nil {
		return spotifeteError
	}
	if trackAdded {
		return UpdateSessionIfNecessary(session)
	}

	if isSessionUpdateNecessary(session, upNextRequest) {
		currentlyPlayingRequest := GetCurrentlyPlayingRequest(session.SimpleListeningSession)
		currentlyPlayingRequest.Status = model.StatusPlayed
		database.GetConnection().Save(currentlyPlayingRequest)

		upNextRequest.Status = model.StatusCurrentlyPlaying
		database.GetConnection().Save(upNextRequest)

		queue := GetSessionQueueInDemocraticOrder(session.SimpleListeningSession)
		if len(queue) > 0 {
			newUpNext := queue[0]
			newUpNext.Status = model.StatusUpNext
			database.GetConnection().Save(&newUpNext)
		}
	}

	return updateSessionPlaylistIfNecessary(session)
}

func isSessionUpdateNecessary(session model.FullListeningSession, upNextRequest *model.SongRequest) bool {
	if upNextRequest == nil {
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

	return isSessionPlaying(session, currentlyPlaying.PlaybackContext) && upNextRequest.SpotifyTrackId == currentlyPlaying.Item.ID.String()
}

func isSessionPlaying(session model.FullListeningSession, playbackContext spotify.PlaybackContext) bool {
	return playbackContext.Type == "playlist" && strings.HasSuffix(string(playbackContext.URI), session.QueuePlaylistId)
}

func updateSessionPlaylistIfNecessary(session model.FullListeningSession) *SpotifeteError {
	currentlyPlayingRequest := GetCurrentlyPlayingRequest(session.SimpleListeningSession)
	upNextRequest := GetUpNextRequest(session.SimpleListeningSession)

	if currentlyPlayingRequest == nil && upNextRequest == nil {
		return nil
	}

	client := Client(session)

	playlist, err := client.GetPlaylist(spotify.ID(session.QueuePlaylistId))
	if err != nil {
		return NewError("Could not get playlist information from Spotify.", err, http.StatusInternalServerError)
	}

	playlistTracks := playlist.Tracks.Tracks

	// First, check playlist length
	if currentlyPlayingRequest != nil && upNextRequest != nil && len(playlistTracks) != 2 {
		return updateSessionPlaylist(session)
	}

	if currentlyPlayingRequest != nil && upNextRequest == nil && len(playlistTracks) != 1 {
		return updateSessionPlaylist(session)
	}

	if currentlyPlayingRequest == nil && upNextRequest == nil && len(playlistTracks) != 0 {
		return updateSessionPlaylist(session)
	}

	// Second, check playlist content
	if currentlyPlayingRequest != nil {
		if playlistTracks[0].Track.ID.String() != currentlyPlayingRequest.SpotifyTrackId {
			return updateSessionPlaylist(session)
		}

		if upNextRequest != nil {
			if playlistTracks[1].Track.ID.String() != upNextRequest.SpotifyTrackId {
				return updateSessionPlaylist(session)
			}
		}
	}

	return nil
}

func updateSessionPlaylist(session model.FullListeningSession) *SpotifeteError {
	client := Client(session)
	currentlyPlayingRequest := GetCurrentlyPlayingRequest(session.SimpleListeningSession)
	upNextRequest := GetUpNextRequest(session.SimpleListeningSession)

	playlistId := spotify.ID(session.QueuePlaylistId)

	// Always replace all tracks with only the current one playing first
	err := client.ReplacePlaylistTracks(playlistId, spotify.ID(currentlyPlayingRequest.SpotifyTrackId))
	if err != nil {
		return NewError("Could not update tracks in playlist.", err, http.StatusInternalServerError)
	}

	// After that, add the up next song as well if it is present
	if upNextRequest != nil {
		_, err = client.AddTracksToPlaylist(playlistId, spotify.ID(upNextRequest.SpotifyTrackId))
		if err != nil {
			return NewError("Could not add track to playlist.", err, http.StatusInternalServerError)
		}
	}

	return nil
}
