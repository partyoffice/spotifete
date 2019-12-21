package dto

import "fmt"

type ListeningSessionDto struct {
	Owner            UserDto            `json:"owner"`
	JoinId           string             `json:"joinId"`
	Title            string             `json:"title"`
	CurrentlyPlaying *TrackMetadataDto  `json:"currentlyPlaying"`
	UpNext           *TrackMetadataDto  `json:"upNext"`
	Queue            []TrackMetadataDto `json:"queue"`
}

func (listeningSessionDto ListeningSessionDto) GetJoinIdHumanReadable() string {
	return fmt.Sprintf("%s %s", listeningSessionDto.JoinId[0:4], listeningSessionDto.JoinId[4:8])
}
