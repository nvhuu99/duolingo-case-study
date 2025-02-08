package test

import (
	"context"
	"time"

	mq "duolingo/lib/message-queue"
	rabbitmq "duolingo/lib/message-queue/driver/rabbitmq"

	"github.com/stretchr/testify/suite"
)

type RabbitMQTopologyTestSuite struct {
    suite.Suite

	Host		string
	Port		string
	User		string
	Password	string

    manager		*rabbitmq.RabbitMQManager
    topology	*rabbitmq.RabbitMQTopology
    managerOpts	*mq.ManagerOptions
	tpOpts		*mq.TopologyOptions
}

func (s *RabbitMQTopologyTestSuite) SetupTest() {
    s.managerOpts = mq.DefaultManagerOptions()
    s.managerOpts.ConnectionTimeOut = 2 * time.Second
    s.managerOpts.HearBeat = 1 * time.Second
    s.manager = rabbitmq.NewRabbitMQManager(context.Background(), s.managerOpts)

    s.tpOpts = mq.DefaultTopologyOptions()
    s.tpOpts.DeclareTimeOut = 2 * time.Second
    s.topology = rabbitmq.NewRabbitMQTopology(context.Background(), s.tpOpts)
    s.topology.UseManager(s.manager)
    s.topology.Topic("mq_tp_test").Queue("mq_q_test").Bind("mq_q_test")
}

func (s *RabbitMQTopologyTestSuite) TearDownTest() {
    s.topology.CleanUp()
    s.manager.Disconnect()
}

func (s *RabbitMQTopologyTestSuite) TestAutoDeclare() {
    s.manager.UseConnection("", "", "", "")
    s.manager.Connect()

    declareErr := s.topology.Declare()
    s.Require().NotNil(declareErr, "topology declare must failed")
    s.Require().False(s.topology.IsReady(), "topology must not be ready")
    s.Assert().Equal(declareErr.Code, mq.DeclareTimeOutExceed ,"declare err code should be DeclareTimeOutExceed")

    s.manager.UseConnection(s.Host, s.Port, s.User, s.Password)
    time.Sleep(s.managerOpts.HearBeat + s.tpOpts.DeclareTimeOut)
    s.Require().True(s.topology.IsReady(), "topology must be declared automatically and successfully")
}
