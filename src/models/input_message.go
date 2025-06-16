package campaign_message

import "encoding/json"

type InputMessage struct {
	Campaign   string `json:"campaign"`
	Title      string `json:"title"`
	Body       string `json:"body"`
	WorkloadId string `json:"workload_id"`
}

func NewInputMessage(campaign string, title string, body string, workloadId string) *InputMessage {
	return &InputMessage{
		Campaign:   campaign,
		Title:      title,
		Body:       body,
		WorkloadId: workloadId,
	}
}

func (m *InputMessage) Encode() []byte {
	marshalled, err := json.Marshal(m)
	if err != nil {
		panic(err)
	}
	return marshalled
}
