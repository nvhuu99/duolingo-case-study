package connection_manager

type ConnectionProxy interface {
	SetArgsPanicIfInvalid(args any)
	GetConnection() (any, error)
	Ping(connection any) error
	IsNetworkErr(err error) bool
	CloseConnection(connection any)
}
