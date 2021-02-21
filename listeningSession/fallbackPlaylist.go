package listeningSession

import (
	"github.com/47-11/spotifete/database"
	"github.com/47-11/spotifete/database/model"
	. "github.com/47-11/spotifete/shared"
	"github.com/47-11/spotifete/users"
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
	fallbackTrackId, spotifeteError := findNextUnplayedFallbackPlaylistTrack(session)
	if spotifeteError != nil {
		return spotifeteError
	}

	_, spotifeteError = RequestSong(session, fallbackTrackId)
	if spotifeteError != nil {
		return spotifeteError
	}

	return nil
}

func findNextUnplayedFallbackPlaylistTrack(session model.FullListeningSession) (nextFallbackTrackId string, spotifeteError *SpotifeteError) {
	return findNextUnplayedFallbackPlaylistTrackOpt(session, 0, 0)
}

func findNextUnplayedFallbackPlaylistTrackOpt(session model.FullListeningSession, maximumPlays uint, pageOffset int) (nextFallbackTrackId string, spotifeteError *SpotifeteError) {
	client := Client(session)
	currentUser, err := client.CurrentUser()
	if err != nil {
		return "", NewError("Could not get user information on session owner from Spotify.", err, http.StatusInternalServerError)
	}

	playlistTracks, err := client.GetPlaylistTracksOpt(spotify.ID(*session.FallbackPlaylistId), &spotify.Options{Offset: &pageOffset, Country: &currentUser.Country}, "")
	if err != nil {
		return "", NewError("Could not get tracks in fallback playlist from Spotify.", err, http.StatusInternalServerError)
	}

	var trackIdsInQueue []string
	database.GetConnection().Select("spotify_track_id").Model(&model.SongRequest{}).Where("session_id = ? and status <> ?", session.ID, model.StatusPlayed).Find(&trackIdsInQueue)

	// TODO: Maybe we could choose a random track? To do that we could just filter all tracks in the current page first and then choose a random one
	for _, playlistTrack := range playlistTracks.Tracks {
		trackId := playlistTrack.Track.ID.String()

		var trackPlays int64
		database.GetConnection().Model(model.SongRequest{}).Where(model.SongRequest{SessionId: session.ID, SpotifyTrackId: trackId}).Count(&trackPlays)

		if trackPlays <= int64(maximumPlays) && !StringSliceContains(trackIdsInQueue, trackId) {
			if isTrackAvailableInUserMarket(*currentUser, playlistTrack.Track) {
				return trackId, nil
			}
		}
	}

	// Nothing found :/
	if len(playlistTracks.Tracks) < playlistTracks.Limit {
		// Checked all playlist tracks -> increase maximum plays and start over
		return findNextUnplayedFallbackPlaylistTrackOpt(session, maximumPlays+1, 0)
	} else {
		// There might still be tracks left that we did not check yet -> increase offset
		return findNextUnplayedFallbackPlaylistTrackOpt(session, maximumPlays, playlistTracks.Offset+playlistTracks.Limit)
	}
}
