package dto

type ListeningSessionDto struct {
	Owner            UserDto            `json:"owner"`
	JoinId           string             `json:"joinId"`
	Title            string             `json:"title"`
	CurrentlyPlaying *TrackMetadataDto  `json:"currentlyPlaying"`
	UpNext           *TrackMetadataDto  `json:"upNext"`
	Queue            []TrackMetadataDto `json:"queue"`
}
