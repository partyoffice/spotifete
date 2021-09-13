package listeningSession

import (
	. "github.com/partyoffice/spotifete/shared"
	. "github.com/partyoffice/spotifete/webapp/apiv2/shared"
)

type NewSessionRequest struct {
	AuthenticatedRequest
	ListeningSessionTitle string `json:"listening_session_title"`
}

func (r NewSessionRequest) Validate() *SpotifeteError {
	if "" == r.ListeningSessionTitle {
		return NewUserError("Missing parameter listening_session_title.")
	}

	return nil
}

type RequestTrackRequest struct {
	TrackId string `json:"track_id"`
}

func (r RequestTrackRequest) Validate() *SpotifeteError {
	if "" == r.TrackId {
		return NewUserError("Missing parameter track_id.")
	}

	return nil
}

type ChangeFallbackPlaylistRequest struct {
	AuthenticatedRequest
	NewFallbackPlaylistId string `json:"new_fallback_playlist_id"`
}

func (r ChangeFallbackPlaylistRequest) Validate() *SpotifeteError {
	if "" == r.NewFallbackPlaylistId {
		return NewUserError("Missing parameter new_fallback_playlist_id.")
	}

	return nil
}

type SetFallbackPlaylistShuffleRequest struct {
	AuthenticatedRequest
	Shuffle bool `json:"shuffle"`
}

type DeleteRequestFromQueueRequest struct {
	AuthenticatedRequest
	SpotifyTrackId string `json:"spotify_track_id""`
}

func (r DeleteRequestFromQueueRequest) Validate() *SpotifeteError {
	if "" == r.SpotifyTrackId {
		return NewUserError("Missing parameter spotify_track_id.")
	}

	return nil
}
