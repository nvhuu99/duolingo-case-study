package metric

import "time"

type Snapshot struct {
	Timestamp time.Time `json:"timestamp" bson:"timestamp"`
	Value     float64 `json:"value" bson:"value"`
	Metadata map[string]string `json:"metadata" bson:"metadata"`
}

func NewSnapshot(val float64, metadata... string) *Snapshot {
	parsed := make(map[string]string)
	if len(metadata) > 0 {
		pair := []string{}
		for _, str := range parsed {
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

	return &Snapshot{
		Timestamp: time.Now(),
		Value: val,
		Metadata: parsed,
	}
}