package test

import (
	"context"
	"time"

	mq "duolingo/lib/message-queue"
	rabbitmq "duolingo/lib/message-queue/driver/rabbitmq"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/suite"
)

type RabbitMQManagerTestSuite struct {
    suite.Suite

    Host		string
	Port		string
	User		string
	Password	string

    manager *rabbitmq.RabbitMQManager
    clients []*ClientMock
    opts    *mq.ManagerOptions
}

func (s *RabbitMQManagerTestSuite) SetupTest() {
    // Manager options
    s.opts = mq.DefaultManagerOptions()
    s.opts.ConnectionTimeOut = 2 * time.Second
    s.opts.HearBeat = 1 * time.Second
    // Create manager
    s.manager = rabbitmq.NewRabbitMQManager(context.Background(), s.opts)
    // Register multiple clients
    for i := 1; i <= 2; i++ {
        client := &ClientMock{}
        client.UseManager(s.manager)
        s.clients = append(s.clients, client)
    }
}

func (s *RabbitMQManagerTestSuite) TearDownTest() {
    s.manager.Disconnect()
}

func (s *RabbitMQManagerTestSuite) TestAutoReconnect() {
    // Firstly, use an invalid connection. After the connection timeout,
    // the connection must failed, and all clients should be informed of the connection failure. 
    s.manager.UseConnection("", "", "", "")
    s.manager.Connect()
    time.Sleep(s.opts.ConnectionTimeOut + s.opts.GraceTimeOut * 2)
    for _, client := range s.clients {
        conn, _ := s.manager.GetClientConnection(client.Id)
        s.Require().Nil(conn, "all client channel must be nil")
        s.Require().True(client.ConnectionFailureTriggered, "manager should call all clients OnConnectionFailure()")
    }

    // Then, switch to a valid connection, and wait for a heartbeat.
    // After this, the manager is expected to reset all clients connection automatically.
    s.manager.UseConnection(s.Host, s.Port, s.User, s.Password)
    time.Sleep(s.opts.HearBeat + s.opts.GraceTimeOut * 2)
    for _, client := range s.clients {
        conn, _ := s.manager.GetClientConnection(client.Id)
        channel, _ := conn.(*amqp.Channel)
        s.Require().NotNil(conn, "all clients channel must be not nil")
        s.Require().False(channel.IsClosed(), "all clients channel must be opened")
        s.Require().True(client.ReConnectedTriggered, "manager should call all clients OnReConnected()")
    }
}
