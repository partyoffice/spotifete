package listeningSession

import (
	"bytes"
	"fmt"
	"github.com/47-11/spotifete/authentication"
	"github.com/47-11/spotifete/config"
	"github.com/47-11/spotifete/database"
	"github.com/47-11/spotifete/database/model"
	. "github.com/47-11/spotifete/error"
	dto "github.com/47-11/spotifete/model/dto"
	"github.com/47-11/spotifete/user"
	"github.com/jinzhu/gorm"
	qrcode "github.com/skip2/go-qrcode"
	"github.com/zmb3/spotify"
	"image/jpeg"
	"math/rand"
	"net/http"
	"time"
)

var numberRunes = []rune("0123456789")

func GetTotalSessionCount() int {
	var count int
	database.GetConnection().Model(&model.ListeningSession{}).Count(&count)
	return count
}

func GetActiveSessionCount() int {
	var count int
	database.GetConnection().Model(&model.ListeningSession{}).Where(model.ListeningSession{Active: true}).Count(&count)
	return count
}

func GetActiveSessions() []model.ListeningSession {
	var sessions []model.ListeningSession
	database.GetConnection().Where(model.ListeningSession{Active: true}).Find(&sessions)
	return sessions
}

func GetSessionById(id uint) *model.ListeningSession {
	var sessions []model.ListeningSession
	database.GetConnection().Where(model.ListeningSession{Model: gorm.Model{ID: id}}).Find(&sessions)

	if len(sessions) == 1 {
		return &sessions[0]
	} else {
		return nil
	}
}

func GetSessionByJoinId(joinId string) *model.ListeningSession {
	if len(joinId) == 0 {
		return nil
	}

	var sessions []model.ListeningSession
	database.GetConnection().Where(model.ListeningSession{JoinId: &joinId}).Find(&sessions)

	if len(sessions) == 1 {
		return &sessions[0]
	} else {
		return nil
	}
}

func GetActiveSessionsByOwnerId(ownerId uint) []model.ListeningSession {
	var sessions []model.ListeningSession
	database.GetConnection().Where(model.ListeningSession{Active: true, OwnerId: ownerId}).Find(&sessions)
	return sessions
}

func GetCurrentlyPlayingRequest(session model.ListeningSession) *model.SongRequest {
	var requests []model.SongRequest
	database.GetConnection().Where(model.SongRequest{
		SessionId: session.ID,
		Status:    model.StatusCurrentlyPlaying,
	}, session.ID).Find(&requests)

	if len(requests) > 0 {
		return &requests[0]
	} else {
		return nil
	}
}

func GetUpNextRequest(session model.ListeningSession) *model.SongRequest {
	var requests []model.SongRequest
	database.GetConnection().Where(model.SongRequest{
		SessionId: session.ID,
		Status:    model.StatusUpNext,
	}, session.ID).Find(&requests)

	if len(requests) > 0 {
		return &requests[0]
	} else {
		return nil
	}
}

func GetSessionQueueInDemocraticOrder(session model.ListeningSession) []model.SongRequest {
	var requests []model.SongRequest
	database.GetConnection().Where(model.SongRequest{
		SessionId: session.ID,
		Status:    model.StatusInQueue,
	}).Order("created_at asc").Find(&requests)

	// TODO: Do something smarter than just using the request order here

	return requests
}

func NewSession(user model.User, title string) (*model.ListeningSession, *SpotifeteError) {
	if len(title) == 0 {
		return nil, NewUserError("Session title must not be empty.")
	}

	client := authentication.GetClientForUser(user)

	joinId := newJoinId()
	playlist, err := client.CreatePlaylistForUser(user.SpotifyId, fmt.Sprintf("%s - SpotiFete", title), fmt.Sprintf("Automatic playlist for SpotiFete session %s. You can join using the code %s-%s or by installing our app and scanning the QR code in the playlist image.", title, joinId[0:4], joinId[4:8]), false)
	if err != nil {
		return nil, NewError("Could not create spotify playlist.", err, http.StatusInternalServerError)
	}

	// Generate QR code for this session
	qrCode, spotifeteError := GenerateQrCodeForSession(joinId, false)
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
	listeningSession := model.ListeningSession{
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

func newJoinId() string {
	for {
		b := make([]rune, 8)
		for i := range b {
			b[i] = numberRunes[rand.Intn(len(numberRunes))]
		}
		newJoinId := string(b)

		if !joinIdExists(newJoinId) {
			return newJoinId
		}
	}
}

func joinIdExists(joinId string) bool {
	var count uint
	database.GetConnection().Model(&model.ListeningSession{}).Where(model.ListeningSession{JoinId: &joinId}).Count(&count)
	return count > 0
}

func CloseSession(user model.User, joinId string) *SpotifeteError {
	session := GetSessionByJoinId(joinId)
	if user.ID != session.OwnerId {
		return NewUserError("Only the session owner can close a session.")
	}

	session.Active = false
	session.JoinId = nil
	database.GetConnection().Save(&session)

	client := authentication.GetClientForUser(user)
	// TODO: Only try to unfollow playlist if owner is still following it.
	err := client.UnfollowPlaylist(spotify.ID(user.SpotifyId), spotify.ID(session.QueuePlaylist))
	if err != nil {
		return NewError("Could not unfollow (delete) playlist.", err, http.StatusInternalServerError)
	}

	// Create rewind playlist if any tracks were requested
	distinctRequestedTracks := GetDistinctRequestedTracks(*session)
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

func IsTrackInQueue(session model.ListeningSession, trackId string) bool {
	var duplicateRequestsForTrack []model.SongRequest
	database.GetConnection().Where("status != 'PLAYED' AND session_id = ? AND spotify_track_id = ?", session.ID, trackId).Find(&duplicateRequestsForTrack)
	return len(duplicateRequestsForTrack) > 0
}

func RequestSong(session model.ListeningSession, trackId string) (model.SongRequest, *SpotifeteError) {
	sessionOwner := user.GetUserById(session.OwnerId)
	client := authentication.GetClientForUser(*sessionOwner)

	// Prevent duplicates
	if IsTrackInQueue(session, trackId) {
		return model.SongRequest{}, NewUserError("This tack is already in the queue.")
	}

	// When using GetTrack Spotify does not include the available markets
	// TODO: Use GetTrack again when Spotify fixed their API
	spotifyTracks, err := client.GetTracks(spotify.ID(trackId))
	if err != nil || len(spotifyTracks) == 0 {
		return model.SongRequest{}, NewError("Could not get track information from Spotify.", err, http.StatusInternalServerError)
	}
	spotifyTrack := spotifyTracks[0]

	updatedTrackMetadata := AddOrUpdateTrackMetadata(*client, *spotifyTrack)

	currentUser, err := client.CurrentUser()
	if err != nil {
		return model.SongRequest{}, NewError("Could not get user information on session owner from Spotify.", err, http.StatusInternalServerError)
	}

	if !isTrackAvailableInUserMarket(*currentUser, *spotifyTrack) {
		return model.SongRequest{}, NewUserError("Sorry, this track is not available :/")
	}

	// Check if we have to add the request to the queue or play it immediately
	currentlyPlayingRequest := GetCurrentlyPlayingRequest(session)
	upNextRequest := GetUpNextRequest(session)

	var newRequestStatus model.SongRequestStatus
	if currentlyPlayingRequest == nil {
		// No song is playing, that means the queue is empty -> Set this to play immediately
		newRequestStatus = model.StatusCurrentlyPlaying
	} else if upNextRequest == nil {
		// A song is currently playing, but no follow up song is present -> Set this as the next song
		newRequestStatus = model.StatusUpNext
	} else {
		// A song is currently playing and a follow up song is present. -> Just add this song to the normal queue
		newRequestStatus = model.StatusInQueue
	}

	newSongRequest := model.SongRequest{
		Model:          gorm.Model{},
		SessionId:      session.ID,
		UserId:         nil,
		SpotifyTrackId: updatedTrackMetadata.SpotifyTrackId,
		Status:         newRequestStatus,
	}

	database.GetConnection().Create(&newSongRequest)

	return newSongRequest, UpdateSessionPlaylistIfNecessary(session)
}

func UpdateSessionIfNecessary(session model.ListeningSession) *SpotifeteError {
	currentlyPlayingRequest := GetCurrentlyPlayingRequest(session)
	upNextRequest := GetUpNextRequest(session)

	owner := user.GetUserById(session.OwnerId)
	client := authentication.GetClientForUser(*owner)
	currentlyPlaying, err := client.PlayerCurrentlyPlaying()
	if err != nil {
		NewInternalError("Could not get currently playing track from Spotify.", err)
		currentlyPlaying = nil
	}

	if session.FallbackPlaylist != nil && upNextRequest == nil {
		// No requests present and a fallback playlist is present
		fallbackTrackId, spotifeteError := findNextUnplayedFallbackPlaylistTrack(session, *client)
		if spotifeteError != nil {
			return spotifeteError
		}

		newSongRequest, spotifeteError := RequestSong(session, fallbackTrackId)
		if spotifeteError != nil {
			return spotifeteError
		}

		switch newSongRequest.Status {
		case model.StatusCurrentlyPlaying:
			currentlyPlayingRequest = &newSongRequest
			break
		case model.StatusUpNext:
			upNextRequest = &newSongRequest
			break
		}
	}

	if currentlyPlaying == nil || currentlyPlaying.Item == nil {
		// Nothing is running -> still update the playlist if neccessary
		return UpdateSessionPlaylistIfNecessary(session)
	}

	currentlyPlayingSpotifyTrackId := currentlyPlaying.Item.ID.String()

	if upNextRequest != nil && upNextRequest.SpotifyTrackId == currentlyPlayingSpotifyTrackId {
		// The previous track finished and the playlist moved on the the next track. Time to update!
		currentlyPlayingRequest.Status = model.StatusPlayed
		database.GetConnection().Save(currentlyPlayingRequest)

		upNextRequest.Status = model.StatusCurrentlyPlaying
		database.GetConnection().Save(upNextRequest)

		queue := GetSessionQueueInDemocraticOrder(session)
		if len(queue) > 0 {
			newUpNext := queue[0]
			newUpNext.Status = model.StatusUpNext
			database.GetConnection().Save(&newUpNext)
		}
	}

	return UpdateSessionPlaylistIfNecessary(session)
}

func findNextUnplayedFallbackPlaylistTrack(session model.ListeningSession, client spotify.Client) (nextFallbackTrackId string, spotifeteError *SpotifeteError) {
	return findNextUnplayedFallbackPlaylistTrackOpt(session, client, 0, 0)
}

func findNextUnplayedFallbackPlaylistTrackOpt(session model.ListeningSession, client spotify.Client, maximumPlays uint, pageOffset int) (nextFallbackTrackId string, spotifeteError *SpotifeteError) {
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
		trackId := playlistTrack.Track.ID.String()

		var trackPlays int
		database.GetConnection().Model(model.SongRequest{}).Where(model.SongRequest{SessionId: session.ID, SpotifyTrackId: trackId}).Count(&trackPlays)

		if uint(trackPlays) <= maximumPlays {
			// Playlist tracks don't include available markets anymore so we have to load the track information explicitly here :/
			// TODO: Remove this if Spotify fixes their API
			refreshedTracks, err := client.GetTracks(playlistTrack.Track.ID)
			if err != nil || len(refreshedTracks) == 0 {
				NewError("Could not fetch track information from Spotify.", err, http.StatusInternalServerError)
			}
			refreshedTrack := refreshedTracks[0]

			if isTrackAvailableInUserMarket(*currentUser, *refreshedTrack) {
				return trackId, nil
			}
		}
	}

	// Nothing found :/
	if len(playlistTracks.Tracks) < playlistTracks.Limit {
		// Checked all playlist tracks -> increase maximum plays and start over
		return findNextUnplayedFallbackPlaylistTrackOpt(session, client, maximumPlays+1, 0)
	} else {
		// There might still be tracks left that we did not check yet -> increase offset
		return findNextUnplayedFallbackPlaylistTrackOpt(session, client, maximumPlays, playlistTracks.Offset+playlistTracks.Limit)
	}
}

func UpdateSessionPlaylistIfNecessary(session model.ListeningSession) *SpotifeteError {
	currentlyPlayingRequest := GetCurrentlyPlayingRequest(session)
	upNextRequest := GetUpNextRequest(session)

	if currentlyPlayingRequest == nil && upNextRequest == nil {
		return nil
	}

	owner := user.GetUserById(session.OwnerId)
	client := authentication.GetClientForUser(*owner)

	playlist, err := client.GetPlaylist(spotify.ID(session.QueuePlaylist))
	if err != nil {
		return NewError("Could not get playlist information from Spotify.", err, http.StatusInternalServerError)
	}

	playlistTracks := playlist.Tracks.Tracks

	// First, check playlist length
	if currentlyPlayingRequest != nil && upNextRequest != nil && len(playlistTracks) != 2 {
		return updateSessionPlaylist(*client, session)
	}

	if currentlyPlayingRequest != nil && upNextRequest == nil && len(playlistTracks) != 1 {
		return updateSessionPlaylist(*client, session)
	}

	if currentlyPlayingRequest == nil && upNextRequest == nil && len(playlistTracks) != 0 {
		return updateSessionPlaylist(*client, session)
	}

	// Second, check playlist content
	if currentlyPlayingRequest != nil {
		if playlistTracks[0].Track.ID.String() != currentlyPlayingRequest.SpotifyTrackId {
			return updateSessionPlaylist(*client, session)
		}

		if upNextRequest != nil {
			if playlistTracks[1].Track.ID.String() != upNextRequest.SpotifyTrackId {
				return updateSessionPlaylist(*client, session)
			}
		}
	}

	return nil
}

func updateSessionPlaylist(client spotify.Client, session model.ListeningSession) *SpotifeteError {
	currentlyPlayingRequest := GetCurrentlyPlayingRequest(session)
	upNextRequest := GetUpNextRequest(session)

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

func PollSessions() {
	for range time.Tick(5 * time.Second) {
		for _, session := range GetActiveSessions() {
			UpdateSessionIfNecessary(session)
		}
	}
}

func CreateDto(listeningSession model.ListeningSession, resolveAdditionalInformation bool) dto.ListeningSessionDto {
	result := dto.ListeningSessionDto{}
	if listeningSession.JoinId == nil {
		result.JoinId = ""
	} else {
		result.JoinId = *listeningSession.JoinId

	}
	result.JoinIdHumanReadable = fmt.Sprintf("%s %s", result.JoinId[0:4], result.JoinId[4:8])
	result.Title = listeningSession.Title

	if resolveAdditionalInformation {
		owner := user.GetUserById(listeningSession.OwnerId)
		result.Owner = user.CreateDto(*owner)
		result.QueuePlaylistId = listeningSession.QueuePlaylist

		if listeningSession.FallbackPlaylist != nil {
			fallbackPlaylist := GetPlaylistMetadataBySpotifyPlaylistId(*listeningSession.FallbackPlaylist)
			fallbackPlaylistDto := dto.PlaylistMetadataDto{}.FromDatabaseModel(*fallbackPlaylist)
			result.FallbackPlaylist = &fallbackPlaylistDto
		}

		currentlyPlayingRequest := GetCurrentlyPlayingRequest(listeningSession)
		if currentlyPlayingRequest != nil {
			currentlyPlayingRequestTrack := dto.TrackMetadataDto{}.FromDatabaseModel(*GetTrackMetadataBySpotifyTrackId(currentlyPlayingRequest.SpotifyTrackId))
			result.CurrentlyPlaying = &currentlyPlayingRequestTrack
		}

		upNextRequest := GetUpNextRequest(listeningSession)
		if upNextRequest != nil {
			upNextRequestTrack := dto.TrackMetadataDto{}.FromDatabaseModel(*GetTrackMetadataBySpotifyTrackId(upNextRequest.SpotifyTrackId))
			result.UpNext = &upNextRequestTrack
		}

		result.Queue = []dto.TrackMetadataDto{}
		for _, request := range GetSessionQueueInDemocraticOrder(listeningSession) {
			requestTrack := GetTrackMetadataBySpotifyTrackId(request.SpotifyTrackId)
			result.Queue = append(result.Queue, dto.TrackMetadataDto{}.FromDatabaseModel(*requestTrack))
		}

		result.QueueLastUpdated = GetQueueLastUpdated(listeningSession)
	}

	return result
}

func GetQueueLastUpdated(session model.ListeningSession) time.Time {
	lastUpdatedSongRequest := model.SongRequest{}
	database.GetConnection().Where(model.SongRequest{
		SessionId: session.ID,
	}).Order("updated_at desc").First(&lastUpdatedSongRequest)

	if lastUpdatedSongRequest.ID != 0 {
		return lastUpdatedSongRequest.UpdatedAt
	} else {
		// No requests found -> Use creation of session
		return session.UpdatedAt
	}
}

func GetDistinctRequestedTracks(session model.ListeningSession) (trackIds []spotify.ID) {
	type Result struct {
		SpotifyTrackId string
	}

	var results []Result
	database.GetConnection().Table("song_requests").Select("distinct spotify_track_id").Where(model.SongRequest{
		SessionId: session.ID,
	}).Scan(&results)

	for _, result := range results {
		trackIds = append(trackIds, spotify.ID(result.SpotifyTrackId))
	}

	return
}

func GenerateQrCodeForSession(joinId string, disableBorder bool) (*qrcode.QRCode, *SpotifeteError) {
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

func ChangeFallbackPlaylist(session model.ListeningSession, user model.User, playlistId string) *SpotifeteError {
	if session.OwnerId != user.ID {
		return NewUserError("Only the session owner can change the fallback playlist.")
	}

	client := authentication.GetClientForUser(user)
	playlistMetadata, err := AddOrUpdatePlaylistMetadata(*client, spotify.ID(playlistId))
	if err != nil {
		return err
	}

	session.FallbackPlaylist = &playlistMetadata.SpotifyPlaylistId
	database.GetConnection().Save(session)

	return nil
}

func isTrackAvailableInUserMarket(user spotify.PrivateUser, track spotify.FullTrack) bool {
	for _, availableMarket := range track.AvailableMarkets {
		if availableMarket == user.Country {
			return true
		}
	}

	return false
}
