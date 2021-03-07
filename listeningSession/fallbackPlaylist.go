package listeningSession

import (
	"fmt"
	"github.com/47-11/spotifete/database"
	"github.com/47-11/spotifete/database/model"
	. "github.com/47-11/spotifete/shared"
	"github.com/47-11/spotifete/users"
	"github.com/google/logger"
	"github.com/zmb3/spotify"
	"net/http"
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

func addFallbackTrackIfNecessary(session model.FullListeningSession, upNextRequest *model.SongRequest) (trackAdded bool, error *SpotifeteError) {
	if session.FallbackPlaylistId == nil {
		return false, nil
	}

	if upNextRequest != nil {
		return false, nil
	}

	spotifeteError := addFallbackTrack(session)
	if spotifeteError != nil {
		return false, spotifeteError
	}

	return true, nil
}

func addFallbackTrack(session model.FullListeningSession) (error *SpotifeteError) {
	fallbackTrackId, spotifeteError := findNextFallbackTrack(session)
	if spotifeteError != nil {
		return spotifeteError
	}

	_, spotifeteError = RequestSong(session, fallbackTrackId)
	if spotifeteError != nil {
		return spotifeteError
	}

	return nil
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
	currentlyPlayingRequest := GetCurrentlyPlayingRequest(session.SimpleListeningSession)

	for i := 0; i < 10_000; i++ {
		fallbackTrack := findPossibleFallbackTrackFromPlayableTracks(playableTracks, session.SimpleListeningSession, currentlyPlayingRequest, 0)
		if fallbackTrack != nil {
			return *fallbackTrack, nil
		}
	}

	session.FallbackPlaylistId = nil
	database.GetConnection().Save(session)

	return "", NewInternalError(fmt.Sprintf("No track found in fallback playlist for session %d that has been played less than 10,000 times. Aborting and removing fallback playlist.", session.ID), nil)
}

func findPossibleFallbackTrackFromPlayableTracks(playableTracks *[]spotify.FullTrack, session model.SimpleListeningSession, currentlyPlayingRequest *model.SongRequest, maximumPlays int64) (possibleFallbackTrackId *string) {
	// TODO: Maybe we could choose a random track? To do that we could just filter all tracks in the current page first and then choose a random one
	for _, playableTrack := range *playableTracks {
		trackId := playableTrack.ID.String()
		if currentlyPlayingRequest == nil || currentlyPlayingRequest.SpotifyTrackId != trackId {
			playCount := getTrackPlayCount(session, trackId)

			if playCount <= maximumPlays {
				return &trackId
			}
		}
	}

	return nil
}
