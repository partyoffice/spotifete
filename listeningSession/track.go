package listeningSession

import (
	"github.com/47-11/spotifete/database"
	"github.com/47-11/spotifete/database/model"
	"github.com/zmb3/spotify"
)

func AddOrUpdateTrackMetadata(spotifyTrack spotify.FullTrack) model.TrackMetadata {
	track := GetTrackMetadataBySpotifyTrackId(spotifyTrack.ID.String())
	if track != nil {
		updatedTrack := track.SetMetadata(spotifyTrack)

		database.GetConnection().Save(&updatedTrack)

		return updatedTrack
	} else {
		newTrack := model.TrackMetadata{}.SetMetadata(spotifyTrack)

		database.GetConnection().Create(&newTrack)

		return newTrack
	}
}

func GetTrackMetadataBySpotifyTrackId(trackId string) *model.TrackMetadata {
	var foundTracks []model.TrackMetadata
	database.GetConnection().Where(model.TrackMetadata{SpotifyTrackId: trackId}).Find(&foundTracks)

	if len(foundTracks) > 0 {
		return &foundTracks[0]
	} else {
		return nil
	}
}

func isTrackAvailableInUserMarket(user spotify.PrivateUser, track spotify.FullTrack) bool {
	for _, availableMarket := range track.Album.AvailableMarkets {
		if availableMarket == user.Country {
			return true
		}
	}

	return false
}
