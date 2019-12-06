package v1

type ErrorResponse struct {
	Message string `json:"message"`
}

type GetAuthUrlResponse struct {
	Url       string `json:"url"`
	SessionId string `json:"sessionId"`
}

type DidAuthSucceedResponse struct {
	Authenticated bool `json:"authenticated"`
}
