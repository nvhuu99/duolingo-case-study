package notification

type Sender interface {
	SendAll(title string, content string, deviceTokens []string) *Result
}
