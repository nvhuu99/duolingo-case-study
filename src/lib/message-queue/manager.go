package messagequeue

type Manager interface {
	GetClientConnection(id string) (any, string)
	RegisterClient(client Client) string
	UnRegisterClient(id string)
}