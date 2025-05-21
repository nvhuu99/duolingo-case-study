package metric

import "time"

type DataPoint struct {
	StartTime  time.Time        `json:"start_time"`
	EndTime    time.Time        `json:"end_time"`
	DurationMs int64           `json:"duration_ms"`
	IncrMs     int64           `json:"incr_ms"`
	Count      int            `json:"count"`
	Snapshots  []*Snapshot `json:"snapshots"`
	Tags      map[string]string `json:"tags" bson:"tags"`
}

func RawDataPoint(snapshots []*Snapshot, tags ...string) *DataPoint {
	parsed := make(map[string]string)
	if len(tags) > 0 {
		pair := []string{}
		for _, str := range tags {
			pair = append(pair, str)
			if len(pair) == 2 {
				parsed[pair[0]] = pair[1]
				pair = []string{}
			}
		}
		if len(pair) == 1 {
			parsed[pair[0]] = ""
		}
	}
	return &DataPoint{
		Snapshots: snapshots,
		Tags: parsed,
	}
}