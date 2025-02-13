package messagequeue

type Manager interface {
	Connect() *Error
	Disconnect()
	GetClientConnection(id string) (any, string)
	RegisterClient(name string, client Client) string
	UnRegisterClient(id string)
	IsReady() bool
}