package listeningSession

import "time"

type QueueLastUpdatedResponse struct {
	QueueLastUpdated time.Time `json:"queue_last_updated"`
}
