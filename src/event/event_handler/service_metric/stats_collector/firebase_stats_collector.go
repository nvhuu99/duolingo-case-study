package metric

import (
	cnst "duolingo/constant"
	firebaseEvt "duolingo/lib/notification/sender/firebase/event"
	"duolingo/lib/metric"
	

	"github.com/google/uuid"
)

type FirebaseStatsCollector struct {
	id    string
	latency int64
	snapshots map[string][]*metric.Snapshot
}

func NewFirebaseStatsCollector() *FirebaseStatsCollector {
	c := new(FirebaseStatsCollector)
	c.id = uuid.NewString()
	c.snapshots = make(map[string][]*metric.Snapshot)
	return c
}

func (c *FirebaseStatsCollector) SubscriberId() string {
	return c.id
}

func (c *FirebaseStatsCollector) Notified(event string, data any) {
	switch event {
	case firebaseEvt.EVT_FIREBASE_SENT_MULTICAST:
		if evt, ok := data.(*firebaseEvt.FirebaseSentMulticastEvent); ok {
			c.latency = max(c.latency, evt.Latency.Milliseconds())
		}
	}
}

func (c *FirebaseStatsCollector) Capture() {
	defer func() {
		c.latency = 0
	}()
	c.snapshots["latency"] = append(c.snapshots["latency"], metric.NewSnapshot(float64(c.latency),
		cnst.METADATA_AGGREGATE_FLAG, "", cnst.METADATA_AGGREGATION_MAXIMUM))
}

func (c *FirebaseStatsCollector) Collect() []*metric.DataPoint {
	defer func() {
		c.snapshots = make(map[string][]*metric.Snapshot)
	}()
	datapoints := []*metric.DataPoint{
		metric.RawDataPoint(c.snapshots["latency"], "metric_target", cnst.METRIC_TARGET_FIREBASE, "metric_name", cnst.METRIC_NAME_MULTICAST_LATENCY),
	}

	return datapoints 
}
