package v1

type InvalidateSessionIdRequest struct {
	SessionId string `json:"sessionId"`
}

type RequestSongRequest struct {
	TrackId string `json:"trackId"`
}
