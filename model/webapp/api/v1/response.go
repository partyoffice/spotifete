package v1

import "github.com/47-11/spotifete/model/dto"

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

type SearchTracksResponse struct {
	Query   string                 `json:"query"`
	Results []dto.TrackMetadataDto `json:"results"`
}
