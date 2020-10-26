package authentication

type NewSessionResponse struct {
	SpotifyAuthenticationUrl string `json:"spotify_authentication_url"`
	SpotifeteSessionId       string `json:"spotifete_session_id"`
}

type IsSessionAuthenticatedResponse struct {
	Authenticated bool `json:"authenticated"`
}
