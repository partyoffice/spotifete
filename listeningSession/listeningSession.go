package listeningSession

import (
	"fmt"
	"github.com/47-11/spotifete/database"
	"github.com/47-11/spotifete/database/model"
	. "github.com/47-11/spotifete/shared"
	"github.com/47-11/spotifete/users"
	"github.com/zmb3/spotify"
	"math/rand"
	"net/http"
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

	if len(listeningSessions) == 1 {
		return &listeningSessions[0]
	} else {
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

	if len(listeningSessions) == 1 {
		return &listeningSessions[0]
	} else {
		return nil
	}
}

func FindFullListeningSessions(filter model.SimpleListeningSession) []model.FullListeningSession {
	var listeningSessions []model.FullListeningSession
	database.GetConnection().Where(filter).Joins("Owner").Joins("FallbackPlaylistMetadata").Find(&listeningSessions)
	return listeningSessions
}

func NewSession(user model.SimpleUser, title string) (*model.SimpleListeningSession, *SpotifeteError) {
	if len(title) == 0 {
		return nil, NewUserError("Session title must not be empty.")
	}

	client := users.Client(user)

	joinId := newJoinId()
	playlist, err := client.CreatePlaylistForUser(user.SpotifyId, fmt.Sprintf("%s - SpotiFete", title), fmt.Sprintf("Automatic playlist for SpotiFete session %s. You can join using the code %s-%s or by installing our app and scanning the QR code in the playlist image.", title, joinId[0:4], joinId[4:8]), false)
	if err != nil {
		return nil, NewError("Could not create spotify playlist.", err, http.StatusInternalServerError)
	}

	qrCode, spotifeteError := QrCodeAsJpeg(joinId, false, 512)
	if spotifeteError != nil {
		return nil, spotifeteError
	}

	go func() {
		err := client.SetPlaylistImage(playlist.ID, qrCode)
		if err != nil {
			NewInternalError("Could not set playlist image.", err)
		}
	}()

	// Create database entry
	listeningSession := model.SimpleListeningSession{
		BaseModel:       model.BaseModel{},
		Active:          true,
		OwnerId:         user.ID,
		JoinId:          &joinId,
		QueuePlaylistId: playlist.ID.String(),
		Title:           title,
	}

	database.GetConnection().Create(&listeningSession)

	return &listeningSession, nil
}

func newJoinId() string {
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
		JoinId: &joinId,
	})

	return existingListeningSession == nil
}

func CloseSession(user model.SimpleUser, joinId string) *SpotifeteError {
	session := FindSimpleListeningSession(model.SimpleListeningSession{
		JoinId: &joinId,
	})
	if session == nil {
		return NewUserError("Unknown listening session.")
	}

	if user.ID != session.OwnerId {
		return NewUserError("Only the session owner can close a session.")
	}

	session.Active = false
	session.JoinId = nil
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

func RequestSong(session model.FullListeningSession, trackId string) (model.SongRequest, *SpotifeteError) {
	client := users.Client(session.Owner)

	// Prevent duplicates
	if IsTrackInQueue(session.SimpleListeningSession, trackId) {
		return model.SongRequest{}, NewUserError("This tack is already in the queue.")
	}

	spotifyTrack, err := client.GetTrack(spotify.ID(trackId))

	updatedTrackMetadata := AddOrUpdateTrackMetadata(*client, *spotifyTrack)

	currentUser, err := client.CurrentUser()
	if err != nil {
		return model.SongRequest{}, NewError("Could not get user information on session owner from Spotify.", err, http.StatusInternalServerError)
	}

	if !isTrackAvailableInUserMarket(*currentUser, *spotifyTrack) {
		return model.SongRequest{}, NewUserError("Sorry, this track is not available :/")
	}

	// Check if we have to add the request to the queue or play it immediately
	currentlyPlayingRequest := GetCurrentlyPlayingRequest(session.SimpleListeningSession)
	upNextRequest := GetUpNextRequest(session.SimpleListeningSession)

	var newRequestStatus model.SongRequestStatus
	if currentlyPlayingRequest == nil {
		// No song is playing, that means the queue is empty -> Set this to play immediately
		newRequestStatus = model.StatusCurrentlyPlaying
	} else if upNextRequest == nil {
		// A song is currently playing, but no follow up song is present -> Set this as the next song
		newRequestStatus = model.StatusUpNext
	} else {
		// A song is currently playing and a follow up song is present. -> Just add this song to the normal queue
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

	return newSongRequest, UpdateSessionPlaylistIfNecessary(session)
}

func UpdateSessionIfNecessary(session model.FullListeningSession) *SpotifeteError {
	currentlyPlayingRequest := GetCurrentlyPlayingRequest(session.SimpleListeningSession)
	upNextRequest := GetUpNextRequest(session.SimpleListeningSession)

	client := users.Client(session.Owner)

	currentlyPlaying, err := client.PlayerCurrentlyPlaying()
	if err != nil {
		NewInternalError("Could not get currently playing track from Spotify.", err)
		currentlyPlaying = nil
	}

	if session.FallbackPlaylistId != nil && upNextRequest == nil {
		// No requests present and a fallback playlist is present
		fallbackTrackId, spotifeteError := findNextUnplayedFallbackPlaylistTrack(session.SimpleListeningSession, *client)
		if spotifeteError != nil {
			return spotifeteError
		}

		newSongRequest, spotifeteError := RequestSong(session, fallbackTrackId)
		if spotifeteError != nil {
			return spotifeteError
		}

		switch newSongRequest.Status {
		case model.StatusCurrentlyPlaying:
			currentlyPlayingRequest = &newSongRequest
			break
		case model.StatusUpNext:
			upNextRequest = &newSongRequest
			break
		}
	}

	if currentlyPlaying == nil || currentlyPlaying.Item == nil {
		// Nothing is running -> still update the playlist if neccessary
		return UpdateSessionPlaylistIfNecessary(session)
	}

	currentlyPlayingSpotifyTrackId := currentlyPlaying.Item.ID.String()

	if upNextRequest != nil && upNextRequest.SpotifyTrackId == currentlyPlayingSpotifyTrackId {
		// The previous track finished and the playlist moved on the the next track. Time to update!
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

	return UpdateSessionPlaylistIfNecessary(session)
}

func UpdateSessionPlaylistIfNecessary(session model.FullListeningSession) *SpotifeteError {
	currentlyPlayingRequest := GetCurrentlyPlayingRequest(session.SimpleListeningSession)
	upNextRequest := GetUpNextRequest(session.SimpleListeningSession)

	if currentlyPlayingRequest == nil && upNextRequest == nil {
		return nil
	}

	client := users.Client(session.Owner)

	playlist, err := client.GetPlaylist(spotify.ID(session.QueuePlaylistId))
	if err != nil {
		return NewError("Could not get playlist information from Spotify.", err, http.StatusInternalServerError)
	}

	playlistTracks := playlist.Tracks.Tracks

	// First, check playlist length
	if currentlyPlayingRequest != nil && upNextRequest != nil && len(playlistTracks) != 2 {
		return updateSessionPlaylist(*client, session.SimpleListeningSession)
	}

	if currentlyPlayingRequest != nil && upNextRequest == nil && len(playlistTracks) != 1 {
		return updateSessionPlaylist(*client, session.SimpleListeningSession)
	}

	if currentlyPlayingRequest == nil && upNextRequest == nil && len(playlistTracks) != 0 {
		return updateSessionPlaylist(*client, session.SimpleListeningSession)
	}

	// Second, check playlist content
	if currentlyPlayingRequest != nil {
		if playlistTracks[0].Track.ID.String() != currentlyPlayingRequest.SpotifyTrackId {
			return updateSessionPlaylist(*client, session.SimpleListeningSession)
		}

		if upNextRequest != nil {
			if playlistTracks[1].Track.ID.String() != upNextRequest.SpotifyTrackId {
				return updateSessionPlaylist(*client, session.SimpleListeningSession)
			}
		}
	}

	return nil
}

func updateSessionPlaylist(client spotify.Client, session model.SimpleListeningSession) *SpotifeteError {
	currentlyPlayingRequest := GetCurrentlyPlayingRequest(session)
	upNextRequest := GetUpNextRequest(session)

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
