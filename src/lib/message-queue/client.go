package messagequeue

type Client interface {
	UseManager(manager Manager)
	NotifyError(chan *Error) chan *Error
	OnConnectionFailure(err *Error)
	OnClientFatalError(err *Error)
	OnReConnected()
}