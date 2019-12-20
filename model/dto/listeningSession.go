package dto

type ListeningSessionDto struct {
	Owner            UserDto            `json:"owner,omitempty"`
	JoinId           string             `json:"joinId"`
	Title            string             `json:"title"`
	CurrentlyPlaying TrackMetadataDto   `json:"currentlyPlaying,omitempty"`
	UpNext           TrackMetadataDto   `json:"upNext,omitempty"`
	Queue            []TrackMetadataDto `json:"queue,omitempty"`
}
