package metric

import "time"

type DataPoint struct {
	StartTime time.Time `json:"start_time"`
	EndTime time.Time `json:"end_time"`
	DurationMs  uint64    `json:"duration_ms"`
	IncrMs uint64 `json:"incr_ms"`
	Count     uint16    `json:"count"`
	Stats map[string][]any     `json:"stats"`
}
