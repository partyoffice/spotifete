package listeningSession

import (
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/google/logger"
	"github.com/partyoffice/spotifete/database"
	"github.com/partyoffice/spotifete/database/model"
	. "github.com/partyoffice/spotifete/shared"
	"github.com/partyoffice/spotifete/users"
	"github.com/zmb3/spotify"
)

func ChangeFallbackPlaylist(session model.SimpleListeningSession, user model.SimpleUser, playlistId string) *SpotifeteError {
	if session.OwnerId != user.ID {
		return NewUserError("Only the session owner can change the fallback playlist.")
	}

	newFallbackPlaylist, err := users.Client(user).GetPlaylist(spotify.ID(playlistId))
	if err != nil {
		return NewError("Could not get playlist information from Spotify.", err, http.StatusInternalServerError)
	}
	playlistMetadata := AddOrUpdatePlaylistMetadata(*newFallbackPlaylist)

	session.FallbackPlaylistId = &playlistMetadata.SpotifyPlaylistId
	database.GetConnection().Save(session)

	return nil
}

func RemoveFallbackPlaylist(session model.SimpleListeningSession, user model.SimpleUser) *SpotifeteError {
	if session.OwnerId != user.ID {
		return NewUserError("Only the session owner can change the fallback playlist.")
	}

	session.FallbackPlaylistId = nil
	database.GetConnection().Save(session)

	return nil
}

func SetFallbackPlaylistShuffle(session model.SimpleListeningSession, user model.SimpleUser, shuffle bool) *SpotifeteError {
	if user.ID != session.OwnerId {
		return NewUserError("Only the session owner can change the shuffle mode.")
	}

	if shuffle == session.FallbackPlaylistShuffle {
		return nil
	}

	session.FallbackPlaylistShuffle = shuffle

	database.GetConnection().Model(&session).Update("fallback_playlist_shuffle", shuffle)

	return nil
}

func addFallbackTrackIfNecessary(session model.FullListeningSession, queue []model.SongRequest) (updatedQueue []model.SongRequest, error *SpotifeteError) {

	if session.FallbackPlaylistId == nil {
		return queue, nil
	}

	for i := len(queue); i < 2; i++ {
		addedRequest, spotifeteError := addFallbackTrack(session)
		if spotifeteError != nil {
			return queue, spotifeteError
		}

		queue = append(queue, addedRequest)
	}

	return queue, nil
}

func addFallbackTrack(session model.FullListeningSession) (addedRequest model.SongRequest, error *SpotifeteError) {

	fallbackTrackId, spotifeteError := findNextFallbackTrack(session)
	if spotifeteError != nil {
		return model.SongRequest{}, spotifeteError
	}

	addedRequest, spotifeteError = RequestSong(session, fallbackTrackId, "Fallback-Playlist")
	if spotifeteError != nil {
		return model.SongRequest{}, spotifeteError
	}

	return addedRequest, nil
}

func findNextFallbackTrack(session model.FullListeningSession) (nextFallbackTrackId string, spotifeteError *SpotifeteError) {

	playableTracks, spotifeteError := getPlayablePlaylistTracks(*session.FallbackPlaylistId, session.Owner)
	if spotifeteError != nil {
		return "", spotifeteError
	}

	if len(*playableTracks) >= 0 {
		return doFindNextFallbackTrack(playableTracks, session)
	}

	session.FallbackPlaylistId = nil
	database.GetConnection().Save(session)

	logger.Info(fmt.Sprintf("Fallback playlist (%s) for session %d does not contain any playable tracks. Removing fallback playlist.",
		*session.FallbackPlaylistId,
		session.ID))
	return "", NewUserError("Fallback playlist does not contain any playable tracks.")
}

func doFindNextFallbackTrack(playableTracks *[]spotify.FullTrack, session model.FullListeningSession) (nextFallbackTrackId string, spotifeteError *SpotifeteError) {

	queue, err := GetFullQueue(session.SimpleListeningSession)
	if err != nil {
		return "", nil
	}

	for i := int64(0); i < 10_000; i++ {
		fallbackTrack, err := findPossibleFallbackTrackFromPlayableTracks(*playableTracks, session.SimpleListeningSession, queue, i)
		if err != nil {
			return "", NewInternalError("could not find possible fallback tracks", err)
		}
		if fallbackTrack != nil {
			return *fallbackTrack, nil
		}
	}

	session.FallbackPlaylistId = nil
	database.GetConnection().Save(session)

	return "", NewInternalError(fmt.Sprintf("No track found in fallback playlist for session %d that has been played less than 10,000 times. Aborting and removing fallback playlist.", session.ID), nil)
}

func findPossibleFallbackTrackFromPlayableTracks(playableTracks []spotify.FullTrack, session model.SimpleListeningSession, queue []model.SongRequest, maximumPlays int64) (possibleFallbackTrackId *string, err error) {
	if session.FallbackPlaylistShuffle {
		rand.Seed(time.Now().UnixNano())
		rand.Shuffle(
			len(playableTracks),
			func(i, j int) {
				playableTracks[i], playableTracks[j] = playableTracks[j], playableTracks[i]
			})
	}

	for _, track := range playableTracks {
		trackId := track.ID.String()
		if !queueContainsTrack(queue, trackId) {
			playCount, err := getTrackPlayCount(session, trackId)
			if err != nil {
				return nil, err
			}

			if playCount <= maximumPlays {
				return &trackId, nil
			}
		}
	}

	return nil, nil
}

func queueContainsTrack(queue []model.SongRequest, trackId string) bool {

	for _, trackInQueue := range queue {
		if trackInQueue.SpotifyTrackId == trackId {
			return true
		}
	}

	return false
}
