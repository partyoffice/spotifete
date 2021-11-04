package listeningSession

import (
	"github.com/partyoffice/spotifete/database/model"
	"github.com/zmb3/spotify"
	"gorm.io/gorm"
)

func AddOrUpdateTrackMetadataInTransaction(spotifyTrack spotify.FullTrack, tx *gorm.DB) (trackMetadata model.TrackMetadata, err error) {

	knownTrackMetadata := GetTrackMetadataBySpotifyTrackIdInTransaction(spotifyTrack.ID.String(), tx)
	if knownTrackMetadata != nil {

		updatedTrackMetadata := knownTrackMetadata.SetMetadata(spotifyTrack)
		err = tx.Save(&updatedTrackMetadata).Error

		return updatedTrackMetadata, err
	} else {

		newTrackMetadata := model.TrackMetadata{}.SetMetadata(spotifyTrack)
		err = tx.Create(&newTrackMetadata).Error

		return newTrackMetadata, err
	}
}

func GetTrackMetadataBySpotifyTrackIdInTransaction(trackId string, tx *gorm.DB) *model.TrackMetadata {

	var foundTracks []model.TrackMetadata
	tx.Where(model.TrackMetadata{SpotifyTrackId: trackId}).Find(&foundTracks)

	if len(foundTracks) > 0 {
		return &foundTracks[0]
	} else {
		return nil
	}
}

func getTrackPlayCount(session model.SimpleListeningSession, spotifyTrackId string) (int64, error) {

	return FindSongRequestCount(model.SongRequest{SessionId: session.ID, SpotifyTrackId: spotifyTrackId})
}
