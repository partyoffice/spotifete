package api

type NewSessionRequest struct {
	LoginSessionId        *string `json:"loginSessionId"`
	ListeningSessionTitle *string `json:"listeningSessionTitle"`
}
