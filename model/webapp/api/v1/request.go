package v1

type InvalidateSessionIdRequest struct {
	LoginSessionId string `json:"sessionId"`
}

type RequestSongRequest struct {
	TrackId string `json:"trackId"`
}

type CreateListeningSessionRequest struct {
	LoginSessionId        *string `json:"loginSessionId"`
	ListeningSessionTitle *string `json:"listeningSessionTitle"`
}

type CloseListeningSessionRequest struct {
	LoginSessionId *string `json:"loginSessionId"`
}
