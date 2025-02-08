package test

import (
	"context"
	"strconv"
	"time"

	mq "duolingo/lib/message-queue"
	rabbitmq "duolingo/lib/message-queue/driver/rabbitmq"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/suite"
)

type RabbitMQTestSuite struct {
    suite.Suite

	Host		string
	Port		string
	User		string
	Password	string

    name        string
    manager		*rabbitmq.RabbitMQManager
    topology	*rabbitmq.RabbitMQTopology
    publisher   *rabbitmq.RabbitMQPublisher
    managerOpts	*mq.ManagerOptions
	tpOpts		*mq.TopologyOptions
    pubOpts     *mq.PublisherOptions
}

func (s *RabbitMQTestSuite) SetupTest() {
    s.name = "rabbitmq_test_" + strconv.Itoa(int(time.Now().UnixMilli())) 

    s.managerOpts = mq.DefaultManagerOptions()
    s.managerOpts.ConnectionTimeOut = 2 * time.Second
    s.managerOpts.HearBeat = 1 * time.Second
    s.manager = rabbitmq.NewRabbitMQManager(context.Background(), s.managerOpts)

    s.tpOpts = mq.DefaultTopologyOptions()
    s.tpOpts.DeclareTimeOut = 2 * time.Second
    s.topology = rabbitmq.NewRabbitMQTopology(context.Background(), s.tpOpts)
    s.topology.UseManager(s.manager)
    s.topology.Topic(s.name).Queue(s.name).Bind(s.name)

    s.pubOpts = mq.DefaultPublisherOptions().
        WithTopic(s.name).
        WithDirectDispatch(s.name).
        WithWriteTimeOut(2 * time.Second)

    s.publisher = rabbitmq.NewPublisher(context.Background(), s.pubOpts)
    s.publisher.UseManager(s.manager)
}

func (s *RabbitMQTestSuite) TearDownTest() {
    s.topology.CleanUp()
    s.manager.Disconnect()
}

func (s *RabbitMQTestSuite) TestManagerAutoReconnect() {
    // Register multiple mock clients
    var clients []*ClientMock
    for i := 1; i <= 2; i++ {
        client := &ClientMock{}
        client.UseManager(s.manager)
        clients = append(clients, client)
    }
    // Firstly, use an invalid connection. After the connection timeout,
    // the connection must failed, and all clients should be informed of the connection failure. 
    s.manager.UseConnection("", "", "", "")
    s.manager.Connect()
    time.Sleep(s.managerOpts.ConnectionTimeOut + s.managerOpts.GraceTimeOut * 2)
    for _, client := range clients {
        conn, _ := s.manager.GetClientConnection(client.Id)
        s.Require().Nil(conn, "all client channel must be nil")
        s.Require().True(client.ConnectionFailureTriggered, "manager should call all clients OnConnectionFailure()")
    }

    // Then, switch to a valid connection, and wait for a heartbeat.
    // After this, the manager is expected to reset all clients connection automatically.
    s.manager.UseConnection(s.Host, s.Port, s.User, s.Password)
    time.Sleep(s.managerOpts.HearBeat + s.managerOpts.GraceTimeOut * 2)
    for _, client := range clients {
        conn, _ := s.manager.GetClientConnection(client.Id)
        channel, _ := conn.(*amqp.Channel)
        s.Require().NotNil(conn, "all clients channel must be not nil")
        s.Require().False(channel.IsClosed(), "all clients channel must be opened")
        s.Require().True(client.ReConnectedTriggered, "manager should call all clients OnReConnected()")
    }
}

func (s *RabbitMQTestSuite) TestTopologyAutoDeclare() {
    s.manager.UseConnection("", "", "", "")
    s.manager.Connect()

    declareErr := s.topology.Declare()
    s.Require().NotNil(declareErr, "topology declare must fail")
    s.Require().False(s.topology.IsReady(), "topology must not be ready")
    s.Assert().Equal(declareErr.Code, mq.DeclareTimeOutExceed, "declare err code should be DeclareTimeOutExceed")

    s.manager.UseConnection(s.Host, s.Port, s.User, s.Password)
    time.Sleep(s.managerOpts.HearBeat + (s.managerOpts.GraceTimeOut * 2) + s.tpOpts.DeclareTimeOut)
    s.Require().True(s.topology.IsReady(), "topology must be declared automatically and successfully")
}

func (s *RabbitMQTestSuite) TestPublisherAutoReconnect() {
    s.manager.UseConnection("", "", "", "")
    s.manager.Connect()

    firstErr := s.publisher.Publish("first message")
    s.Require().NotNil(firstErr, "first message must fail to be published")
    s.Assert().Equal(firstErr.Code, mq.PublishTimeOutExceed, "pub err code should be PublishTimeOutExceed")

    s.manager.UseConnection(s.Host, s.Port, s.User, s.Password)
    time.Sleep(s.managerOpts.HearBeat + (s.managerOpts.GraceTimeOut * 2) + s.tpOpts.DeclareTimeOut)
    
    secErr := s.publisher.Publish("second message")
    s.Require().Nil(secErr, "second message must be published successfully")
    thirdErr := s.publisher.Publish("third message")
    s.Require().Nil(thirdErr, "second message must be published successfully")
}

