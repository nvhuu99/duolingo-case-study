package test

import (
	mq "duolingo/lib/message-queue"

	"github.com/stretchr/testify/mock"
)

type ClientMock struct {
	mock.Mock

	Id                         string
	ConnectionFailureTriggered bool
	ReConnectedTriggered       bool
	ClientFatalErrorTriggered  bool

	manager mq.Manager
	errChan chan error
}

func (client *ClientMock) UseManager(manager mq.Manager) {
	client.Id = manager.RegisterClient("client mock", client)
	client.manager = manager
}

func (client *ClientMock) NotifyError(ch chan error) chan error {
	client.errChan = ch
	return ch
}

func (client *ClientMock) OnConnectionFailure(err error) {
	client.ConnectionFailureTriggered = true
}

func (client *ClientMock) OnClientFatalError(err error) {
	client.ClientFatalErrorTriggered = true
}

func (client *ClientMock) OnReConnected() {
	client.ReConnectedTriggered = true
}
