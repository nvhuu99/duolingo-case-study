package message_queue

type Client interface {
	UseManager(manager Manager)
	ResetConnection()
}
