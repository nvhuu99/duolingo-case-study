package connection_manager

type ConnectionProxy interface {
	SetConnectionArgsWithPanicOnValidationErr(args any)
	CreateConnection() (any, error)
	Ping(connection any) error
	IsNetworkError(err error) bool
	CloseConnection(connection any)
}
