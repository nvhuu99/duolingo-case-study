package messagequeue

type Client interface {
	UseManager(manager Manager)
	NotifyError(chan error) chan error
	OnConnectionFailure(err error)
	OnClientFatalError(err error)
	OnReConnected()
}
