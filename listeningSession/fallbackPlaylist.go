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

	client := users.Client(user)
	playlistMetadata, err := AddOrUpdatePlaylistMetadata(*client, spotify.ID(playlistId))
	if err != nil {
		return err
	}

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

func findNextUnplayedFallbackPlaylistTrack(session model.SimpleListeningSession, client spotify.Client) (nextFallbackTrackId string, spotifeteError *SpotifeteError) {
	return findNextUnplayedFallbackPlaylistTrackOpt(session, client, 0, 0)
}

func findNextUnplayedFallbackPlaylistTrackOpt(session model.SimpleListeningSession, client spotify.Client, maximumPlays uint, pageOffset int) (nextFallbackTrackId string, spotifeteError *SpotifeteError) {
	currentUser, err := client.CurrentUser()
	if err != nil {
		return "", NewError("Could not get user information on session owner from Spotify.", err, http.StatusInternalServerError)
	}

	playlistTracks, err := client.GetPlaylistTracksOpt(spotify.ID(*session.FallbackPlaylistId), &spotify.Options{Offset: &pageOffset, Country: &currentUser.Country}, "")
	if err != nil {
		return "", NewError("Could not get tracks in fallback playlist from Spotify.", err, http.StatusInternalServerError)
	}

	// TODO: Maybe we could choose a random track? To do that we could just filter all tracks in the current page first and then choose a random one
	for _, playlistTrack := range playlistTracks.Tracks {
		trackId := playlistTrack.Track.ID.String()

		var trackPlays int64
		database.GetConnection().Model(model.SongRequest{}).Where(model.SongRequest{SessionId: session.ID, SpotifyTrackId: trackId}).Count(&trackPlays)

		if trackPlays <= int64(maximumPlays) {
			// Playlist tracks don't include available markets anymore so we have to load the track information explicitly here :/
			// TODO: Remove this if Spotify fixes their API
			refreshedTracks, err := client.GetTracks(playlistTrack.Track.ID)
			if err != nil || len(refreshedTracks) == 0 {
				NewError("Could not fetch track information from Spotify.", err, http.StatusInternalServerError)
			}
			refreshedTrack := refreshedTracks[0]

			if isTrackAvailableInUserMarket(*currentUser, *refreshedTrack) {
				return trackId, nil
			}
		}
	}

	// Nothing found :/
	if len(playlistTracks.Tracks) < playlistTracks.Limit {
		// Checked all playlist tracks -> increase maximum plays and start over
		return findNextUnplayedFallbackPlaylistTrackOpt(session, client, maximumPlays+1, 0)
	} else {
		// There might still be tracks left that we did not check yet -> increase offset
		return findNextUnplayedFallbackPlaylistTrackOpt(session, client, maximumPlays, playlistTracks.Offset+playlistTracks.Limit)
	}
}
