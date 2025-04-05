package test

import (
	"context"
	"strings"
	// "strconv"
	"time"

	mq "duolingo/lib/message-queue"
	rabbitmq "duolingo/lib/message-queue/driver/rabbitmq"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/suite"
)

var (
	graceTimeOut   = 200 * time.Millisecond
	connTimeOut    = 1 * time.Second
	declareTimeOut = 1 * time.Second
	writeTimeOut   = 1 * time.Second
	heartBeat      = 1 * time.Second
)

type RabbitMQTestSuite struct {
	suite.Suite

	Host     string
	Port     string
	User     string
	Password string

	name string

	manager   *rabbitmq.RabbitMQManager
	topology  *rabbitmq.RabbitMQTopology
	publisher *rabbitmq.RabbitMQPublisher
	consumer  *rabbitmq.RabbitMQConsumer
}

func (s *RabbitMQTestSuite) SetupTest() {
	s.name = "rabbitmq_driver_test_" // + strconv.Itoa(int(time.Now().UnixMilli()))
	s.manager = rabbitmq.NewRabbitMQManager(context.Background())
	s.topology = rabbitmq.NewRabbitMQTopology("topology", context.Background())
	s.publisher = rabbitmq.NewPublisher("publisher", context.Background())
	s.consumer = rabbitmq.NewConsumer("consumer", context.Background())

	s.manager.
		WithOptions(nil).
		WithGraceTimeOut(graceTimeOut).
		WithConnectionTimeOut(connTimeOut).
		WithHearBeat(heartBeat).
		WithKeepAlive(true)

	s.topology.
		UseManager(s.manager)
	s.topology.
		WithOptions(nil).
		WithGraceTimeOut(graceTimeOut).
		WithDeclareTimeOut(declareTimeOut).
		WithQueuesPurged(true)
	s.topology.
		Topic(s.name).Queue(s.name).Bind(s.name)

	s.publisher.
		UseManager(s.manager)
	s.publisher.
		WithOptions(nil).
		WithGraceTimeOut(graceTimeOut).
		WithWriteTimeOut(writeTimeOut).
		WithTopic(s.name).
		WithDirectDispatch(s.name)

	s.consumer.
		UseManager(s.manager)
	s.consumer.
		WithOptions(nil).
		WithGraceTimeOut(graceTimeOut).
		WithQueue(s.name)
}

func (s *RabbitMQTestSuite) TearDownTest() {
	s.topology.CleanUp()
	s.manager.Disconnect()
}

func (s *RabbitMQTestSuite) WaitForConnection(wait time.Duration, tick time.Duration) {
	timeOut := time.After(wait)
	for {
		select {
		case <-timeOut:
			return
		default:
			if !s.manager.IsReady() {
				time.Sleep(tick)
				continue
			}
			if !s.topology.IsReady() {
				time.Sleep(tick)
				continue
			}
			return
		}
	}
}

func (s *RabbitMQTestSuite) TestManagerAutoReconnect() {
	var clients [2]*ClientMock
	for i := 0; i < 2; i++ {
		client := new(ClientMock)
		client.UseManager(s.manager)
		clients[i] = client
	}

	s.manager.UseConnection("", "", "", "")
	s.manager.Connect()
	time.Sleep(connTimeOut + 2 * graceTimeOut)

	for _, client := range clients {
		conn, _ := s.manager.GetClientConnection(client.Id)
		s.Require().Nil(conn, "all client channel must be nil")
		s.Require().True(client.ConnectionFailureTriggered, "manager should call all clients OnConnectionFailure()")
	}

	s.manager.UseConnection(s.Host, s.Port, s.User, s.Password)
	s.WaitForConnection(10 * time.Second, graceTimeOut)
	for _, client := range clients {
		conn, _ := client.manager.GetClientConnection(client.Id)
		channel, _ := conn.(*amqp.Channel)
		s.Require().NotNil(conn, "all clients channel must be not nil")
		s.Require().False(channel.IsClosed(), "all clients channel must be opened")
		s.Require().True(client.ReConnectedTriggered, "manager should call all clients OnReConnected()")
	}
}

func (s *RabbitMQTestSuite) TestTopologyAutoDeclare() {
	s.manager.UseConnection("", "", "", "")
	s.manager.Connect()
	time.Sleep(connTimeOut + graceTimeOut * 2)

	declareErr := s.topology.Declare()
	s.Require().NotNil(declareErr, "topology declare must fail")
	s.Require().False(s.topology.IsReady(), "topology must not be ready")
	s.Assert().True(strings.HasPrefix(declareErr.Error(), mq.ErrMessages[mq.ERR_DECLARE_TIMEOUT_EXCEED]), "must be ERR_DECLARE_TIMEOUT_EXCEED")

	s.manager.UseConnection(s.Host, s.Port, s.User, s.Password)
	s.WaitForConnection(10*time.Second, graceTimeOut)
	s.Require().True(s.topology.IsReady(), "topology must be declared automatically and successfully")
}

func (s *RabbitMQTestSuite) TestPublisherAutoReconnect() {
	s.manager.UseConnection("", "", "", "")
	s.manager.Connect()
	time.Sleep(connTimeOut + graceTimeOut * 2)

	firstErr := s.publisher.Publish("first message")
	s.Require().NotNil(firstErr, "first message must fail to be published")
	s.Assert().True(strings.HasPrefix(firstErr.Error(), mq.ErrMessages[mq.ERR_PUBLISH_TIMEOUT_EXCEED]), "should be ERR_PUBLISH_TIMEOUT_EXCEED")

	s.manager.UseConnection(s.Host, s.Port, s.User, s.Password)
	s.WaitForConnection(10*time.Second, graceTimeOut)

	secErr := s.publisher.Publish("second message")
	s.Require().Nil(secErr, "second message must be published successfully")
	thirdErr := s.publisher.Publish("third message")
	s.Require().Nil(thirdErr, "second message must be published successfully")
}

func (s *RabbitMQTestSuite) TestConsumerAutoReconnect() {
	s.manager.UseConnection("", "", "", "")
	s.manager.Connect()
	time.Sleep(connTimeOut + graceTimeOut * 2)

	go func() {
		time.Sleep(time.Second)
		s.manager.UseConnection(s.Host, s.Port, s.User, s.Password)
		s.WaitForConnection(10 * time.Second, graceTimeOut)
		s.publisher.Publish("sample")
	}()

	done := make(chan bool, 1)
	s.consumer.Consume(done, func(message string) mq.ConsumerAction {
		defer func() {
			done <- true
		}()
		s.Require().Equal(message, "sample", "consumer must receive the correct message")
		return mq.ConsumerAccept
	})
}

func (s *RabbitMQTestSuite) TestConsumerActionAccept() {
	s.manager.UseConnection(s.Host, s.Port, s.User, s.Password)
	s.manager.Connect()
	s.WaitForConnection(10*time.Second, graceTimeOut)
	
	messages := [2]string{"first", "second"}
	for _, mssg := range messages {
		s.publisher.Publish(mssg)
	}

	index := 0
	done := make(chan bool, 1)
	result := false
	s.consumer.Consume(done, func(mssg string) mq.ConsumerAction {
		defer func() {
			if !result || index == len(messages) {
				done <- true
			}
		}()
		result = s.Assert().Equal(mssg, messages[index], "consumer must receive messages in the correct order")
		result = true
		index++
		return mq.ConsumerAccept
	})
}

func (s *RabbitMQTestSuite) TestConsumerActionRequeue() {
	s.manager.UseConnection(s.Host, s.Port, s.User, s.Password)
	s.manager.Connect()
	s.WaitForConnection(10*time.Second, graceTimeOut)
	
	s.publisher.Publish("sample")
	
	first := true
	done := make(chan bool, 1)
	s.consumer.Consume(done, func(mssg string) mq.ConsumerAction {
		if first {
			first = false
			return mq.ConsumerRequeue
		} else {
			s.Assert().Equal(mssg, "sample", "consumer must receive the same message after it is requeued")
			done <- true
			return mq.ConsumerAccept
		}
	})
}

func (s *RabbitMQTestSuite) TestConsumerActionReject() {
	s.manager.UseConnection(s.Host, s.Port, s.User, s.Password)
	s.manager.Connect()
	s.WaitForConnection(10*time.Second, graceTimeOut)

	messages := [2]string{"first", "second"}
	for _, mssg := range messages {
		s.publisher.Publish(mssg)
	}

	index := 0
	done := make(chan bool, 1)
	result := false
	s.consumer.Consume(done, func(mssg string) mq.ConsumerAction {
		defer func() {
			if !result || index == len(messages) {
				done <- true
			}
		}()
		result = s.Assert().Equal(mssg, messages[index], "consumer must receive messages in the correct order")
		result = true
		index++
		return mq.ConsumerReject
	})
}

func (s *RabbitMQTestSuite) TestConsumerConfirmationRecovery() {
    s.manager.UseConnection(s.Host, s.Port, s.User, s.Password)
    s.manager.Connect()
    s.WaitForConnection(10 * time.Second, graceTimeOut)
    
	s.publisher.Publish("first")

    isFirst := true
    done := make(chan bool, 1)
    s.consumer.Consume(done, func (message string) mq.ConsumerAction {
        if isFirst {
            s.manager.UseConnection("", "", "", "")
            time.Sleep(connTimeOut + graceTimeOut * 2)

            go func() {
                s.manager.UseConnection(s.Host, s.Port, s.User, s.Password)
                s.WaitForConnection(10 * time.Second, graceTimeOut)

                time.Sleep(graceTimeOut * 2)
                s.publisher.Publish("second")
            }()

            isFirst = false

            return mq.ConsumerAccept
        } else {
            s.Require().Equal(message, "second", "the first message should be automatically acknowleged, and the consumer should receive the second message")
            done <- true
            return mq.ConsumerAccept
        }
    })
}
