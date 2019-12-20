package service

import (
	"errors"
	"fmt"
	"github.com/47-11/spotifete/database"
	. "github.com/47-11/spotifete/model/database"
	dto "github.com/47-11/spotifete/model/dto"
	"github.com/getsentry/sentry-go"
	"github.com/jinzhu/gorm"
	"github.com/zmb3/spotify"
	"log"
	"math/rand"
	"sync"
	"time"
)

type listeningSessionService struct {
	numberRunes []rune
}

var listeningSessionServiceInstance *listeningSessionService
var listeningSessionServiceOnce sync.Once

func ListeningSessionService() *listeningSessionService {
	listeningSessionServiceOnce.Do(func() {
		listeningSessionServiceInstance = &listeningSessionService{
			numberRunes: []rune("0123456789"),
		}
	})
	return listeningSessionServiceInstance
}

func (listeningSessionService) GetTotalSessionCount() int {
	var count int
	database.Connection.Model(&ListeningSession{}).Count(&count)
	return count
}

func (listeningSessionService) GetActiveSessionCount() int {
	var count int
	database.Connection.Model(&ListeningSession{}).Where(ListeningSession{Active: true}).Count(&count)
	return count
}

func (listeningSessionService) GetActiveSessions() []ListeningSession {
	var sessions []ListeningSession
	database.Connection.Where(ListeningSession{Active: true}).Find(&sessions)
	return sessions
}

func (listeningSessionService) GetSessionById(id uint) *ListeningSession {
	var sessions []ListeningSession
	database.Connection.Where(ListeningSession{Model: gorm.Model{ID: id}}).Find(&sessions)

	if len(sessions) == 1 {
		return &sessions[0]
	} else {
		return nil
	}
}

func (listeningSessionService) GetSessionByJoinId(joinId string) *ListeningSession {
	if len(joinId) == 0 {
		return nil
	}

	var sessions []ListeningSession
	database.Connection.Where(ListeningSession{JoinId: &joinId}).Find(&sessions)

	if len(sessions) == 1 {
		return &sessions[0]
	} else {
		return nil
	}
}

func (listeningSessionService) GetActiveSessionsByOwnerId(ownerId uint) []ListeningSession {
	var sessions []ListeningSession
	database.Connection.Where(ListeningSession{Active: true, OwnerId: ownerId}).Find(&sessions)
	return sessions
}

func (s listeningSessionService) GetCurrentlyPlayingAndUpNext(session ListeningSession) (currentlyPlaying *SongRequest, upNext *SongRequest, err error) {
	// Get currently playing
	var currentlyPlayingResults []SongRequest
	database.Connection.Where("session_id = ? AND status = 'CURRENTLY_PLAYING'", session.ID).Find(&currentlyPlayingResults)

	switch len(currentlyPlayingResults) {
	case 0:
		currentlyPlaying = nil
		break
	case 1:
		currentlyPlaying = &currentlyPlayingResults[0]
		break
	default:
		return nil, nil, errors.New(fmt.Sprintf("found invalid number of request with status CURRENTLY_PLAYING: %d", len(currentlyPlayingResults)))
	}

	// Get up next
	var upNextResults []SongRequest
	database.Connection.Where("session_id = ? AND status = 'UP_NEXT'", session.ID).Find(&upNextResults)

	switch len(upNextResults) {
	case 0:
		upNext = nil
		break
	case 1:
		upNext = &upNextResults[0]
		break
	default:
		return nil, nil, errors.New(fmt.Sprintf("found invalid number of request with status UP_NEXT: %d", len(upNextResults)))
	}

	return currentlyPlaying, upNext, nil
}

func (s listeningSessionService) GetSessionQueueInDemocraticOrder(session ListeningSession) []SongRequest {
	var requests []SongRequest
	database.Connection.Where(SongRequest{
		SessionId: session.ID,
		Status:    IN_QUEUE,
	}).Find(&requests)

	// TODO: Do something smart here

	return requests
}

func (s listeningSessionService) NewSession(user *User, title string) (*ListeningSession, error) {
	client := SpotifyService().GetAuthenticator().NewClient(user.GetToken())

	joinId := s.newJoinId()
	playlist, err := client.CreatePlaylistForUser(user.SpotifyId, fmt.Sprintf("%s - SpotiFete", title), fmt.Sprintf("Automatic playlist for SpotiFete session %s. You can join using the code %s %s.", title, joinId[0:4], joinId[4:8]), false)
	if err != nil {
		return nil, err
	}

	listeningSession := ListeningSession{
		Model:           gorm.Model{},
		Active:          true,
		OwnerId:         user.ID,
		JoinId:          &joinId,
		SpotifyPlaylist: playlist.ID.String(),
		Title:           title,
	}

	database.Connection.Create(&listeningSession)

	return &listeningSession, nil
}

func (s listeningSessionService) newJoinId() string {
	for {
		b := make([]rune, 8)
		for i := range b {
			b[i] = s.numberRunes[rand.Intn(len(s.numberRunes))]
		}
		newJoinId := string(b)

		if !s.joinIdExists(newJoinId) {
			return newJoinId
		}
	}
}

func (listeningSessionService) joinIdExists(joinId string) bool {
	var count uint
	database.Connection.Model(&ListeningSession{}).Where(ListeningSession{JoinId: &joinId}).Count(&count)
	return count > 0
}

func (s listeningSessionService) CloseSession(user *User, joinId string) error {
	session := s.GetSessionByJoinId(joinId)
	if user.ID != session.OwnerId {
		return errors.New("only the owner can close a session")
	}

	session.Active = false
	session.JoinId = nil
	database.Connection.Save(&session)

	client := SpotifyService().authenticator.NewClient(user.GetToken())
	return client.UnfollowPlaylist(spotify.ID(user.SpotifyId), spotify.ID(session.SpotifyPlaylist))
}

func (s listeningSessionService) RequestSong(session *ListeningSession, trackId string) error {
	sessionOwner := UserService().GetUserById(session.OwnerId)
	client := SpotifyService().GetAuthenticator().NewClient(sessionOwner.GetToken())

	// Prevent duplicates
	trackMetadata := SpotifyService().GetTrackMetadataBySpotifyTrackId(trackId)
	if trackMetadata != nil {
		var duplicateRequestsForTrack []SongRequest
		database.Connection.Where("status != 'PLAYED' AND session_id = ? AND track_id = ?", session.ID, trackMetadata.ID).Find(&duplicateRequestsForTrack)
		if len(duplicateRequestsForTrack) > 0 {
			return errors.New("that song is already in the queue")
		}
	}

	// Check if we have to add the request to the queue or play it immediately
	currentlyPlayingRequest, upNextRequest, err := s.GetCurrentlyPlayingAndUpNext(*session)
	if err != nil {
		return err
	}

	var newRequestStatus SongRequestStatus
	if currentlyPlayingRequest == nil {
		// No song is playing, that means the queue is empty -> Set this to play immediately
		newRequestStatus = CURRENTLY_PLAYING
	} else if upNextRequest == nil {
		// A song is currently playing, but no follow up song is present -> Set this as the next song
		newRequestStatus = UP_NEXT
	} else {
		// A song is currently playing and a follow up song is present. -> Just add this song to the normal queue
		newRequestStatus = IN_QUEUE
	}

	updatedTrackMetadata, err := SpotifyService().AddOrUpdateTrackMetadata(client, spotify.ID(trackId))
	if err != nil {
		return err
	}

	newSongRequest := SongRequest{
		Model:     gorm.Model{},
		SessionId: session.ID,
		UserId:    nil,
		TrackId:   updatedTrackMetadata.ID,
		Status:    newRequestStatus,
	}

	database.Connection.Create(&newSongRequest)

	return nil
}

func (s listeningSessionService) UpdateSessionIfNeccessary(session ListeningSession) error {
	currentlyPlayingRequest, upNextRequest, err := s.GetCurrentlyPlayingAndUpNext(session)
	if err != nil {
		return err
	}

	owner := UserService().GetUserById(session.OwnerId)
	client := SpotifyService().GetAuthenticator().NewClient(owner.GetToken())
	currentlyPlaying, err := client.PlayerCurrentlyPlaying()
	if err != nil {
		return err
	}

	if currentlyPlaying == nil || currentlyPlaying.Item == nil {
		// Nothing is running -> do nothing
		return nil
	}

	currentlyPlayingSpotifyTrackId := currentlyPlaying.Item.ID.String()

	if currentlyPlayingRequest == nil {
		// No requests present
		// TODO: A this point we could use a fallback playlist or replay previously played tracks from this session
		return nil
	}

	currentlyPlayingRequestTrack := SpotifyService().GetTrackMetadataById(currentlyPlayingRequest.TrackId)
	if currentlyPlayingRequestTrack.SpotifyTrackId == currentlyPlayingSpotifyTrackId {
		// The current track is still in progress -> NO-OP
	}

	upNextRequestTrack := SpotifyService().GetTrackMetadataById(upNextRequest.TrackId)
	if upNextRequest != nil && upNextRequestTrack.SpotifyTrackId == currentlyPlayingSpotifyTrackId {
		// The previous track finished and the playlist moved on the the next track. Time to update!
		currentlyPlayingRequest.Status = PLAYED
		database.Connection.Save(currentlyPlayingRequest)

		upNextRequest.Status = CURRENTLY_PLAYING
		database.Connection.Save(upNextRequest)

		queue := s.GetSessionQueueInDemocraticOrder(session)
		if len(queue) > 0 {
			newUpNext := queue[0]
			newUpNext.Status = UP_NEXT
			database.Connection.Save(&newUpNext)
		}
	}

	return s.UpdateSessionPlaylistIfNeccessary(session)
}

func (s listeningSessionService) UpdateSessionPlaylistIfNeccessary(session ListeningSession) error {
	currentlyPlayingRequest, upNextRequest, err := s.GetCurrentlyPlayingAndUpNext(session)
	if err != nil {
		return err
	}

	if currentlyPlayingRequest == nil && upNextRequest == nil {
		return nil
	}

	owner := UserService().GetUserById(session.OwnerId)
	client := SpotifyService().GetAuthenticator().NewClient(owner.GetToken())

	playlist, err := client.GetPlaylist(spotify.ID(session.SpotifyPlaylist))
	if err != nil {
		return err
	}

	playlistTracks := playlist.Tracks.Tracks

	// First, check playlist length
	if currentlyPlayingRequest != nil && upNextRequest != nil && len(playlistTracks) != 2 {
		return s.updateSessionPlaylist(client, session)
	}

	if currentlyPlayingRequest != nil && upNextRequest == nil && len(playlistTracks) != 1 {
		return s.updateSessionPlaylist(client, session)
	}

	if currentlyPlayingRequest == nil && upNextRequest == nil && len(playlistTracks) != 0 {
		return s.updateSessionPlaylist(client, session)
	}

	// Second, check playlist content
	if currentlyPlayingRequest != nil {
		currentlyPlayingRequestTrack := SpotifyService().GetTrackMetadataById(currentlyPlayingRequest.TrackId)
		if playlistTracks[0].Track.ID.String() != currentlyPlayingRequestTrack.SpotifyTrackId {
			return s.updateSessionPlaylist(client, session)
		}

		if upNextRequest != nil {
			upNextRequestTrack := SpotifyService().GetTrackMetadataById(upNextRequest.TrackId)
			if playlistTracks[1].Track.ID.String() != upNextRequestTrack.SpotifyTrackId {
				return s.updateSessionPlaylist(client, session)
			}
		}
	}

	return nil
}

func (s listeningSessionService) updateSessionPlaylist(client spotify.Client, session ListeningSession) error {
	currentlyPlayingRequest, upNextRequest, err := s.GetCurrentlyPlayingAndUpNext(session)
	if err != nil {
		return err
	}

	playlistId := spotify.ID(session.SpotifyPlaylist)

	// Always replace all tracks with only the current one playing first
	currentlyPlayingRequestTrack := SpotifyService().GetTrackMetadataById(currentlyPlayingRequest.TrackId)
	err = client.ReplacePlaylistTracks(playlistId, spotify.ID(currentlyPlayingRequestTrack.SpotifyTrackId))
	if err != nil {
		return err
	}

	// After that, add the up next song as well if it is present
	if upNextRequest != nil {
		upNextRequestTrack := SpotifyService().GetTrackMetadataById(upNextRequest.TrackId)
		_, err = client.AddTracksToPlaylist(playlistId, spotify.ID(upNextRequestTrack.SpotifyTrackId))
		if err != nil {
			return err
		}
	}

	return nil
}

func (s listeningSessionService) PollSessions() {
	for _ = range time.Tick(5 * time.Second) {
		for _, session := range s.GetActiveSessions() {
			err := s.UpdateSessionIfNeccessary(session)
			if err != nil {
				log.Println(err)
				sentry.CaptureException(err)
			}
		}
	}
}

func (s listeningSessionService) CreateDto(listeningSession ListeningSession, resolveAdditionalInformation bool) dto.ListeningSessionDto {
	result := dto.ListeningSessionDto{}
	result.JoinId = *listeningSession.JoinId
	result.Title = listeningSession.Title

	if resolveAdditionalInformation {
		owner := UserService().GetUserById(listeningSession.OwnerId)
		result.Owner = UserService().CreateDto(*owner, false)

		currentlyPlayingRequest, upNextRequest, err := s.GetCurrentlyPlayingAndUpNext(listeningSession)
		if err != nil {
			panic(err)
		}

		if currentlyPlayingRequest != nil {
			currentlyPlayingRequestTrack := SpotifyService().GetTrackMetadataById(currentlyPlayingRequest.TrackId)
			result.CurrentlyPlaying = dto.TrackMetadataDto{}.FromDatabaseModel(*currentlyPlayingRequestTrack)
		}

		if upNextRequest != nil {
			upNextRequestTrack := SpotifyService().GetTrackMetadataById(upNextRequest.TrackId)
			result.UpNext = dto.TrackMetadataDto{}.FromDatabaseModel(*upNextRequestTrack)
		}

		result.Queue = []dto.TrackMetadataDto{}
		for _, request := range s.GetSessionQueueInDemocraticOrder(listeningSession) {
			requestTrack := SpotifyService().GetTrackMetadataById(request.TrackId)
			result.Queue = append(result.Queue, dto.TrackMetadataDto{}.FromDatabaseModel(*requestTrack))
		}
	}

	return result
}
