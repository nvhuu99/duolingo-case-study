package notification

type Sender interface {
	SendAll(message *Message, deviceTokens []string) (*Result, error)
}