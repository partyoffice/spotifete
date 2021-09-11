package listeningSession

import (
	"fmt"
	"net/http"
	"time"

	"github.com/partyoffice/spotifete/database"
	"github.com/partyoffice/spotifete/database/model"
	. "github.com/partyoffice/spotifete/shared"
	"github.com/partyoffice/spotifete/users"
	"github.com/patrickmn/go-cache"
	"github.com/zmb3/spotify"
)

var playlistTrackCache = cache.New(30*time.Minute, time.Hour)

func AddOrUpdatePlaylistMetadata(playlist spotify.FullPlaylist) model.PlaylistMetadata {
	knownPlaylistMetadata := GetPlaylistMetadataBySpotifyPlaylistId(playlist.ID.String())
	if knownPlaylistMetadata != nil {
		updatedPlaylistMetadata := knownPlaylistMetadata.FromFullPlaylist(playlist)

		database.GetConnection().Save(&updatedPlaylistMetadata)

		return updatedPlaylistMetadata
	} else {
		newPlaylistMetadata := model.PlaylistMetadata{}.FromFullPlaylist(playlist)

		database.GetConnection().Create(&newPlaylistMetadata)

		return newPlaylistMetadata
	}
}

func GetPlaylistMetadataBySpotifyPlaylistId(playlistId string) *model.PlaylistMetadata {
	var foundPlaylists []model.PlaylistMetadata
	database.GetConnection().Where(model.PlaylistMetadata{SpotifyPlaylistId: playlistId}).Find(&foundPlaylists)

	if len(foundPlaylists) > 0 {
		return &foundPlaylists[0]
	} else {
		return nil
	}
}

func createPlaylistForSession(joinId string, sessionTitle string, user model.SimpleUser) (*spotify.FullPlaylist, *SpotifeteError) {

	client := users.Client(user)

	playlistTitle := fmt.Sprintf("%s - SpotiFete", sessionTitle)
	playlistDescription := fmt.Sprintf("Automatic playlist for Spotifete session %s. You can join using the code %s-%s or by installing our app and scanning the QR code in the playlist image.", playlistTitle, joinId[0:4], joinId[4:8])

	playlist, err := client.CreatePlaylistForUser(user.SpotifyId, playlistTitle, playlistDescription, false)
	if err != nil {
		return nil, NewError("Could not create spotify playlist.", err, http.StatusInternalServerError)
	}

	go setPlaylistImage(playlist, joinId, user)
	return playlist, nil
}

func setPlaylistImage(playlist *spotify.FullPlaylist, joinId string, user model.SimpleUser) *SpotifeteError {

	client := users.Client(user)

	qrCode, spotifeteError := QrCodeAsJpeg(joinId, false, 512)
	if spotifeteError != nil {
		return spotifeteError
	}

	err := client.SetPlaylistImage(playlist.ID, qrCode)
	if err == nil {
		return nil
	} else {
		return NewError("Could not create spotify playlist.", err, http.StatusInternalServerError)
	}
}

func getPlayablePlaylistTracks(playlistId string, user model.SimpleUser) (*[]spotify.FullTrack, *SpotifeteError) {
	cachedTracks, found := playlistTrackCache.Get(playlistId)
	if found {
		return cachedTracks.(*[]spotify.FullTrack), nil
	}

	loadedTracks, spotifeteError := loadPlaylistTracksFromSpotify(playlistId, user)
	if spotifeteError != nil {
		return nil, spotifeteError
	}

	var playableTracks []spotify.FullTrack
	for _, playlistTrack := range loadedTracks {
		track := playlistTrack.Track

		if track.IsPlayable != nil && *track.IsPlayable {
			playableTracks = append(playableTracks, track)
		}
	}

	playlistTrackCache.SetDefault(playlistId, &playableTracks)
	return &playableTracks, nil
}

func loadPlaylistTracksFromSpotify(playlistId string, user model.SimpleUser) ([]spotify.PlaylistTrack, *SpotifeteError) {
	client := users.Client(user)

	spotifyPlaylistId := spotify.ID(playlistId)
	searchOptions := spotify.Options{Country: &user.Country}
	var tracks []spotify.PlaylistTrack

	var tracksLeftToLoad = true
	for tracksLeftToLoad {
		page, err := client.GetPlaylistTracksOpt(spotifyPlaylistId, &searchOptions, "")
		if err != nil {
			return nil, NewError("Could not get playlist tracks from Spotify.", err, http.StatusInternalServerError)
		}

		tracks = append(tracks, page.Tracks...)

		loadedTrackCount := len(tracks)
		if loadedTrackCount >= page.Total {
			tracksLeftToLoad = false
		} else {
			searchOptions.Offset = &loadedTrackCount
		}
	}

	return tracks, nil
}
