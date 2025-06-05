package campaign_message

type InputMessage struct {
	campaign   string
	title      string
	body       string
	workloadId string
}

func NewInputMessage(campaign string, title string, body string, workloadId string) *InputMessage {
	return &InputMessage{
		campaign:   campaign,
		title:      title,
		body:       body,
		workloadId: workloadId,
	}
}
