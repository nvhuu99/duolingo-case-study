package event

import "time"

const (
	EVT_MONGODB_QUERY = "evt_mongodb_query"
)

type MongoDBQueryEvent struct {
	Latency time.Duration
}
