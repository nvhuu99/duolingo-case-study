package messagequeue

type Manager interface {
	Connect()
	Disconnect()
	GetClientConnection(id string) (any, string)
	RegisterClient(client Client) string
	UnRegisterClient(id string)
}