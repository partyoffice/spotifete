package listeningSession

import (
	"github.com/47-11/spotifete/database"
	"github.com/47-11/spotifete/database/model"
	. "github.com/47-11/spotifete/shared"
	"github.com/47-11/spotifete/users"
	"github.com/patrickmn/go-cache"
	"github.com/zmb3/spotify"
	"net/http"
	"time"
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
