package connection_manager

type ConnectionProxy interface {
	CreateConnection(args *ConnectArgs) (any, error)
	Ping(connection any) error
	IsNetworkError(err error) bool
	CloseConnection(connection any)
}
