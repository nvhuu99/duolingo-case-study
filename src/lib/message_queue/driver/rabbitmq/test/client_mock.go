package test

import (
	mq "duolingo/lib/message_queue"

	"github.com/stretchr/testify/mock"
)

type ClientMock struct {
	mock.Mock

	Id                   string
	ReConnectedTriggered bool

	manager mq.Manager
}

func (client *ClientMock) UseManager(manager mq.Manager) {
	client.Id = manager.RegisterClient("client mock", client)
	client.manager = manager
}

func (client *ClientMock) ResetConnection() {
	client.ReConnectedTriggered = true
}
