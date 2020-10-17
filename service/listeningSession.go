package service

import (
	"bytes"
	"fmt"
	"github.com/47-11/spotifete/config"
	"github.com/47-11/spotifete/database"
	. "github.com/47-11/spotifete/error"
	. "github.com/47-11/spotifete/model/database"
	dto "github.com/47-11/spotifete/model/dto"
	"github.com/jinzhu/gorm"
	qrcode "github.com/skip2/go-qrcode"
	"github.com/zmb3/spotify"
	"image/jpeg"
	"math/rand"
	"net/http"
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
	database.GetConnection().Model(&ListeningSession{}).Count(&count)
	return count
}

func (listeningSessionService) GetActiveSessionCount() int {
	var count int
	database.GetConnection().Model(&ListeningSession{}).Where(ListeningSession{Active: true}).Count(&count)
	return count
}

func (listeningSessionService) GetActiveSessions() []ListeningSession {
	var sessions []ListeningSession
	database.GetConnection().Where(ListeningSession{Active: true}).Find(&sessions)
	return sessions
}

func (listeningSessionService) GetSessionById(id uint) *ListeningSession {
	var sessions []ListeningSession
	database.GetConnection().Where(ListeningSession{Model: gorm.Model{ID: id}}).Find(&sessions)

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
	database.GetConnection().Where(ListeningSession{JoinId: &joinId}).Find(&sessions)

	if len(sessions) == 1 {
		return &sessions[0]
	} else {
		return nil
	}
}

func (listeningSessionService) GetActiveSessionsByOwnerId(ownerId uint) []ListeningSession {
	var sessions []ListeningSession
	database.GetConnection().Where(ListeningSession{Active: true, OwnerId: ownerId}).Find(&sessions)
	return sessions
}

func (s listeningSessionService) GetCurrentlyPlayingRequest(session ListeningSession) *SongRequest {
	var requests []SongRequest
	database.GetConnection().Where(SongRequest{
		SessionId: session.ID,
		Status:    StatusCurrentlyPlaying,
	}, session.ID).Find(&requests)

	if len(requests) > 0 {
		return &requests[0]
	} else {
		return nil
	}
}

func (s listeningSessionService) GetUpNextRequest(session ListeningSession) *SongRequest {
	var requests []SongRequest
	database.GetConnection().Where(SongRequest{
		SessionId: session.ID,
		Status:    StatusUpNext,
	}, session.ID).Find(&requests)

	if len(requests) > 0 {
		return &requests[0]
	} else {
		return nil
	}
}

func (s listeningSessionService) GetSessionQueueInDemocraticOrder(session ListeningSession) []SongRequest {
	var requests []SongRequest
	database.GetConnection().Where(SongRequest{
		SessionId: session.ID,
		Status:    StatusInQueue,
	}).Order("created_at asc").Find(&requests)

	// TODO: Do something smarter than just using the request order here

	return requests
}

func (s listeningSessionService) NewSession(user User, title string) (*ListeningSession, *SpotifeteError) {
	if len(title) == 0 {
		return nil, NewUserError("Session title must not be empty.")
	}

	client := SpotifyService().GetClientForUser(user)

	joinId := s.newJoinId()
	playlist, err := client.CreatePlaylistForUser(user.SpotifyId, fmt.Sprintf("%s - SpotiFete", title), fmt.Sprintf("Automatic playlist for SpotiFete session %s. You can join using the code %s-%s or by installing our app and scanning the QR code in the playlist image.", title, joinId[0:4], joinId[4:8]), false)
	if err != nil {
		return nil, NewError("Could not create spotify playlist.", err, http.StatusInternalServerError)
	}

	// Generate QR code for this session
	qrCode, spotifeteError := s.GenerateQrCodeForSession(joinId, false)
	if spotifeteError != nil {
		return nil, spotifeteError
	}

	// Encode QR code as jpeg
	jpegBuffer := new(bytes.Buffer)
	err = jpeg.Encode(jpegBuffer, qrCode.Image(512), nil)
	if err != nil {
		return nil, NewError("Could not encode qr code as image.", err, http.StatusInternalServerError)
	}

	// Set QR code as playlist image in background
	go func() {
		err := client.SetPlaylistImage(playlist.ID, jpegBuffer)
		if err != nil {
			NewInternalError("Could not set playlist image.", err)
		}
	}()

	// Create database entry
	listeningSession := ListeningSession{
		Model:         gorm.Model{},
		Active:        true,
		OwnerId:       user.ID,
		JoinId:        &joinId,
		QueuePlaylist: playlist.ID.String(),
		Title:         title,
	}

	database.GetConnection().Create(&listeningSession)

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
	database.GetConnection().Model(&ListeningSession{}).Where(ListeningSession{JoinId: &joinId}).Count(&count)
	return count > 0
}

func (s listeningSessionService) CloseSession(user User, joinId string) *SpotifeteError {
	session := s.GetSessionByJoinId(joinId)
	if user.ID != session.OwnerId {
		return NewUserError("Only the session owner can close a session.")
	}

	session.Active = false
	session.JoinId = nil
	database.GetConnection().Save(&session)

	client := SpotifyService().GetClientForUser(user)
	// TODO: Only try to unfollow playlist if owner is still following it.
	err := client.UnfollowPlaylist(spotify.ID(user.SpotifyId), spotify.ID(session.QueuePlaylist))
	if err != nil {
		return NewError("Could not unfollow (delete) playlist.", err, http.StatusInternalServerError)
	}

	// Create rewind playlist if any tracks were requested
	distinctRequestedTracks := s.GetDistinctRequestedTracks(*session)
	if len(distinctRequestedTracks) > 0 {
		rewindPlaylist, err := client.CreatePlaylistForUser(user.SpotifyId, fmt.Sprintf("%s Rewind - SpotiFete", session.Title), fmt.Sprintf("Rewind playlist for your session %s. This contains all the songs that were requested.", session.Title), false)
		if err != nil {
			return NewError("Could not create rewind playlist.", err, http.StatusInternalServerError)
		}

		go func() {
			var page []spotify.ID
			for _, track := range distinctRequestedTracks {
				page = append(page, track)

				if len(page) == 100 {
					_, err = client.AddTracksToPlaylist(rewindPlaylist.ID, page...)
					if err != nil {
						_ = NewInternalError(fmt.Sprintf("Could not add tracks to rewind playlist. Session: %d | Rewind playlist: %s | Tracks: %s", session.ID, rewindPlaylist.ID, page), err)
					}
					page = []spotify.ID{}
				}
			}

			if len(page) > 0 {
				_, err = client.AddTracksToPlaylist(rewindPlaylist.ID, page...)
				if err != nil {
					_ = NewInternalError(fmt.Sprintf("Could not add tracks to rewind playlist. Session: %d | Rewind playlist: %s | Tracks: %s", session.ID, rewindPlaylist.ID, page), err)
				}
			}
		}()
	}

	return nil
}

func (s listeningSessionService) IsTrackInQueue(session ListeningSession, trackId string) bool {
	var duplicateRequestsForTrack []SongRequest
	database.GetConnection().Where("status != 'PLAYED' AND session_id = ? AND spotify_track_id = ?", session.ID, trackId).Find(&duplicateRequestsForTrack)
	return len(duplicateRequestsForTrack) > 0
}

func (s listeningSessionService) RequestSong(session ListeningSession, trackId string) *SpotifeteError {
	sessionOwner := UserService().GetUserById(session.OwnerId)
	client := SpotifyService().GetClientForUser(*sessionOwner)

	// Prevent duplicates
	if s.IsTrackInQueue(session, trackId) {
		return NewUserError("This tack is already in the queue.")
	}

	// When using GetTrack Spotify does not include the available markets
	// TODO: Use GetTrack again when Spotify fixed their API
	spotifyTracks, err := client.GetTracks(spotify.ID(trackId))
	if err != nil || len(spotifyTracks) == 0 {
		return NewError("Could not get track information from Spotify.", err, http.StatusInternalServerError)
	}
	spotifyTrack := spotifyTracks[0]

	updatedTrackMetadata := SpotifyService().AddOrUpdateTrackMetadata(*client, *spotifyTrack)

	currentUser, err := client.CurrentUser()
	if err != nil {
		return NewError("Could not get user information on session owner from Spotify.", err, http.StatusInternalServerError)
	}

	if !s.isTrackAvailableInUserMarket(*currentUser, *spotifyTrack) {
		return NewUserError("Sorry, this track is not available :/")
	}

	// Check if we have to add the request to the queue or play it immediately
	currentlyPlayingRequest := s.GetCurrentlyPlayingRequest(session)
	upNextRequest := s.GetUpNextRequest(session)

	var newRequestStatus SongRequestStatus
	if currentlyPlayingRequest == nil {
		// No song is playing, that means the queue is empty -> Set this to play immediately
		newRequestStatus = StatusCurrentlyPlaying
	} else if upNextRequest == nil {
		// A song is currently playing, but no follow up song is present -> Set this as the next song
		newRequestStatus = StatusUpNext
	} else {
		// A song is currently playing and a follow up song is present. -> Just add this song to the normal queue
		newRequestStatus = StatusInQueue
	}

	newSongRequest := SongRequest{
		Model:          gorm.Model{},
		SessionId:      session.ID,
		UserId:         nil,
		SpotifyTrackId: updatedTrackMetadata.SpotifyTrackId,
		Status:         newRequestStatus,
	}

	database.GetConnection().Create(&newSongRequest)

	return s.UpdateSessionPlaylistIfNecessary(session)
}

func (s listeningSessionService) UpdateSessionIfNecessary(session ListeningSession) *SpotifeteError {
	currentlyPlayingRequest := s.GetCurrentlyPlayingRequest(session)
	upNextRequest := s.GetUpNextRequest(session)

	owner := UserService().GetUserById(session.OwnerId)
	client := SpotifyService().GetClientForUser(*owner)
	currentlyPlaying, err := client.PlayerCurrentlyPlaying()
	if err != nil {
		NewInternalError("Could not get currently playing track from Spotify.", err)
		currentlyPlaying = nil
	}

	if currentlyPlaying == nil || currentlyPlaying.Item == nil {
		// Nothing is running -> still update the playlist if neccessary
		return s.UpdateSessionPlaylistIfNecessary(session)
	}

	currentlyPlayingSpotifyTrackId := currentlyPlaying.Item.ID.String()

	if session.FallbackPlaylist != nil && upNextRequest == nil {
		// No requests present and a fallback playlist is present
		fallbackTrackId, spotifeteError := s.findNextUnplayedFallbackPlaylistTrack(session, *client)
		if spotifeteError != nil {
			return spotifeteError
		}

		spotifeteError = s.RequestSong(session, fallbackTrackId)
		if spotifeteError != nil {
			return spotifeteError
		}

		return nil
	}

	if upNextRequest != nil && upNextRequest.SpotifyTrackId == currentlyPlayingSpotifyTrackId {
		// The previous track finished and the playlist moved on the the next track. Time to update!
		currentlyPlayingRequest.Status = StatusPlayed
		database.GetConnection().Save(currentlyPlayingRequest)

		upNextRequest.Status = StatusCurrentlyPlaying
		database.GetConnection().Save(upNextRequest)

		queue := s.GetSessionQueueInDemocraticOrder(session)
		if len(queue) > 0 {
			newUpNext := queue[0]
			newUpNext.Status = StatusUpNext
			database.GetConnection().Save(&newUpNext)
		}
	}

	return s.UpdateSessionPlaylistIfNecessary(session)
}

func (s listeningSessionService) findNextUnplayedFallbackPlaylistTrack(session ListeningSession, client spotify.Client) (nextFallbackTrackId string, spotifeteError *SpotifeteError) {
	return s.findNextUnplayedFallbackPlaylistTrackOpt(session, client, 0, 0)
}

func (s listeningSessionService) findNextUnplayedFallbackPlaylistTrackOpt(session ListeningSession, client spotify.Client, maximumPlays uint, pageOffset int) (nextFallbackTrackId string, spotifeteError *SpotifeteError) {
	currentUser, err := client.CurrentUser()
	if err != nil {
		return "", NewError("Could not get user information on session owner from Spotify.", err, http.StatusInternalServerError)
	}

	playlistTracks, err := client.GetPlaylistTracksOpt(spotify.ID(*session.FallbackPlaylist), &spotify.Options{Offset: &pageOffset, Country: &currentUser.Country}, "")
	if err != nil {
		return "", NewError("Could not get tracks in fallback playlist from Spotify.", err, http.StatusInternalServerError)
	}

	// TODO: Maybe we could choose a random track? To do that we could just filter all tracks in the current page first and then choose a random one
	for _, playlistTrack := range playlistTracks.Tracks {
		// Playlist tracks don't include available markets anymore so we have to load the track information explicitly here :/
		// TODO: Remove this
		refreshedTracks, err := client.GetTracks(playlistTrack.Track.ID)
		if err != nil || len(refreshedTracks) == 0 {
			NewError("Could not fetch track information from Spotify.", err, http.StatusInternalServerError)
		}
		track := refreshedTracks[0]
		trackId := track.ID.String()

		var trackPlays int
		database.GetConnection().Model(SongRequest{}).Where(SongRequest{SessionId: session.ID, SpotifyTrackId: trackId}).Count(&trackPlays)

		if uint(trackPlays) <= maximumPlays && s.isTrackAvailableInUserMarket(*currentUser, *track) {
			return trackId, nil
		}
	}

	// Nothing found :/
	if len(playlistTracks.Tracks) < playlistTracks.Limit {
		// Checked all playlist tracks -> increase maximum plays and start over
		return s.findNextUnplayedFallbackPlaylistTrackOpt(session, client, maximumPlays+1, 0)
	} else {
		// There might still be tracks left that we did not check yet -> increase offset
		return s.findNextUnplayedFallbackPlaylistTrackOpt(session, client, maximumPlays, playlistTracks.Offset+playlistTracks.Limit)
	}
}

func (s listeningSessionService) UpdateSessionPlaylistIfNecessary(session ListeningSession) *SpotifeteError {
	currentlyPlayingRequest := s.GetCurrentlyPlayingRequest(session)
	upNextRequest := s.GetUpNextRequest(session)

	if currentlyPlayingRequest == nil && upNextRequest == nil {
		return nil
	}

	owner := UserService().GetUserById(session.OwnerId)
	client := SpotifyService().GetClientForUser(*owner)

	playlist, err := client.GetPlaylist(spotify.ID(session.QueuePlaylist))
	if err != nil {
		return NewError("Could not get playlist information from Spotify.", err, http.StatusInternalServerError)
	}

	playlistTracks := playlist.Tracks.Tracks

	// First, check playlist length
	if currentlyPlayingRequest != nil && upNextRequest != nil && len(playlistTracks) != 2 {
		return s.updateSessionPlaylist(*client, session)
	}

	if currentlyPlayingRequest != nil && upNextRequest == nil && len(playlistTracks) != 1 {
		return s.updateSessionPlaylist(*client, session)
	}

	if currentlyPlayingRequest == nil && upNextRequest == nil && len(playlistTracks) != 0 {
		return s.updateSessionPlaylist(*client, session)
	}

	// Second, check playlist content
	if currentlyPlayingRequest != nil {
		if playlistTracks[0].Track.ID.String() != currentlyPlayingRequest.SpotifyTrackId {
			return s.updateSessionPlaylist(*client, session)
		}

		if upNextRequest != nil {
			if playlistTracks[1].Track.ID.String() != upNextRequest.SpotifyTrackId {
				return s.updateSessionPlaylist(*client, session)
			}
		}
	}

	return nil
}

func (s listeningSessionService) updateSessionPlaylist(client spotify.Client, session ListeningSession) *SpotifeteError {
	currentlyPlayingRequest := s.GetCurrentlyPlayingRequest(session)
	upNextRequest := s.GetUpNextRequest(session)

	playlistId := spotify.ID(session.QueuePlaylist)

	// Always replace all tracks with only the current one playing first
	err := client.ReplacePlaylistTracks(playlistId, spotify.ID(currentlyPlayingRequest.SpotifyTrackId))
	if err != nil {
		return NewError("Could not update tracks in playlist.", err, http.StatusInternalServerError)
	}

	// After that, add the up next song as well if it is present
	if upNextRequest != nil {
		_, err = client.AddTracksToPlaylist(playlistId, spotify.ID(upNextRequest.SpotifyTrackId))
		if err != nil {
			return NewError("Could not add track to playlist.", err, http.StatusInternalServerError)
		}
	}

	return nil
}

func (s listeningSessionService) PollSessions() {
	for range time.Tick(5 * time.Second) {
		for _, session := range s.GetActiveSessions() {
			s.UpdateSessionIfNecessary(session)
		}
	}
}

func (s listeningSessionService) CreateDto(listeningSession ListeningSession, resolveAdditionalInformation bool) dto.ListeningSessionDto {
	result := dto.ListeningSessionDto{}
	if listeningSession.JoinId == nil {
		result.JoinId = ""
	} else {
		result.JoinId = *listeningSession.JoinId

	}
	result.JoinIdHumanReadable = fmt.Sprintf("%s %s", result.JoinId[0:4], result.JoinId[4:8])
	result.Title = listeningSession.Title

	if resolveAdditionalInformation {
		owner := UserService().GetUserById(listeningSession.OwnerId)
		result.Owner = UserService().CreateDto(*owner, false)
		result.QueuePlaylistId = listeningSession.QueuePlaylist

		if listeningSession.FallbackPlaylist != nil {
			fallbackPlaylist := SpotifyService().GetPlaylistMetadataBySpotifyPlaylistId(*listeningSession.FallbackPlaylist)
			fallbackPlaylistDto := dto.PlaylistMetadataDto{}.FromDatabaseModel(*fallbackPlaylist)
			result.FallbackPlaylist = &fallbackPlaylistDto
		}

		currentlyPlayingRequest := s.GetCurrentlyPlayingRequest(listeningSession)
		if currentlyPlayingRequest != nil {
			currentlyPlayingRequestTrack := dto.TrackMetadataDto{}.FromDatabaseModel(*SpotifyService().GetTrackMetadataBySpotifyTrackId(currentlyPlayingRequest.SpotifyTrackId))
			result.CurrentlyPlaying = &currentlyPlayingRequestTrack
		}

		upNextRequest := s.GetUpNextRequest(listeningSession)
		if upNextRequest != nil {
			upNextRequestTrack := dto.TrackMetadataDto{}.FromDatabaseModel(*SpotifyService().GetTrackMetadataBySpotifyTrackId(upNextRequest.SpotifyTrackId))
			result.UpNext = &upNextRequestTrack
		}

		result.Queue = []dto.TrackMetadataDto{}
		for _, request := range s.GetSessionQueueInDemocraticOrder(listeningSession) {
			requestTrack := SpotifyService().GetTrackMetadataBySpotifyTrackId(request.SpotifyTrackId)
			result.Queue = append(result.Queue, dto.TrackMetadataDto{}.FromDatabaseModel(*requestTrack))
		}

		result.QueueLastUpdated = s.GetQueueLastUpdated(listeningSession)
	}

	return result
}

func (s listeningSessionService) GetQueueLastUpdated(session ListeningSession) time.Time {
	lastUpdatedSongRequest := SongRequest{}
	database.GetConnection().Where(SongRequest{
		SessionId: session.ID,
	}).Order("updated_at desc").First(&lastUpdatedSongRequest)

	if lastUpdatedSongRequest.ID != 0 {
		return lastUpdatedSongRequest.UpdatedAt
	} else {
		// No requests found -> Use creation of session
		return session.UpdatedAt
	}
}

func (s listeningSessionService) GetDistinctRequestedTracks(session ListeningSession) (trackIds []spotify.ID) {
	type Result struct {
		SpotifyTrackId string
	}

	var results []Result
	database.GetConnection().Table("song_requests").Select("distinct spotify_track_id").Where(SongRequest{
		SessionId: session.ID,
	}).Scan(&results)

	for _, result := range results {
		trackIds = append(trackIds, spotify.ID(result.SpotifyTrackId))
	}

	return
}

func (listeningSessionService) GenerateQrCodeForSession(joinId string, disableBorder bool) (*qrcode.QRCode, *SpotifeteError) {
	baseUrl := config.Get().SpotifeteConfiguration.BaseUrl
	qrCodeContent := fmt.Sprintf("%s/session/view/%s", baseUrl, joinId)

	// Generate QR code for this session
	qrCode, err := qrcode.New(qrCodeContent, qrcode.Medium)
	if err != nil {
		return nil, NewError("Could not create QR code.", err, http.StatusInternalServerError)
	}

	qrCode.DisableBorder = disableBorder
	return qrCode, nil
}

func (s listeningSessionService) ChangeFallbackPlaylist(session ListeningSession, user User, playlistId string) *SpotifeteError {
	if session.OwnerId != user.ID {
		return NewUserError("Only the session owner can change the fallback playlist.")
	}

	client := SpotifyService().GetClientForUser(user)
	playlistMetadata, err := SpotifyService().AddOrUpdatePlaylistMetadata(*client, spotify.ID(playlistId))
	if err != nil {
		return err
	}

	session.FallbackPlaylist = &playlistMetadata.SpotifyPlaylistId
	database.GetConnection().Save(session)

	return nil
}

func (s listeningSessionService) isTrackAvailableInUserMarket(user spotify.PrivateUser, track spotify.FullTrack) bool {
	for _, availableMarket := range track.AvailableMarkets {
		if availableMarket == user.Country {
			return true
		}
	}

	return false
}
