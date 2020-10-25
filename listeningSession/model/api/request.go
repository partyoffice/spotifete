package api

type NewListeningSessionRequest struct {
	LoginSessionId        *string `json:"loginSessionId"`
	ListeningSessionTitle *string `json:"listeningSessionTitle"`
}
