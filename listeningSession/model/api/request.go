package api

type NewSessionRequest struct {
	LoginSessionId        *string `json:"loginSessionId"`
	ListeningSessionTitle *string `json:"listeningSessionTitle"`
}

type CloseSessionRequest struct {
	LoginSessionId *string `json:"loginSessionId"`
}
