package event

import "time"

const (
	EVT_FIREBASE_SENT_MULTICAST = "evt_firebase_sent_multicast"
)

type FirebaseSentMulticastEvent struct {
	Latency time.Duration
}
