// Warning: DO NOT run the test on production environment or all messages will be purged.
// You should only run it on test environments.
package test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/stretchr/testify/suite"

	mq "duolingo/lib/message-queue"
	rabbitmq "duolingo/lib/message-queue/driver/rabbitmq"
	"duolingo/model"
)

var (
	graceTimeOut   = 200 * time.Millisecond
	connTimeOut    = 1 * time.Second
	declareTimeOut = 1 * time.Second
	heartBeat      = 1 * time.Second
)

type MessageInputApiTestSuite struct {
	suite.Suite

	Host     string
	Port     string
	User     string
	Password string

	manager   *rabbitmq.RabbitMQManager
	consumer  *rabbitmq.RabbitMQConsumer
	topology  *rabbitmq.RabbitMQTopology
}

func (s *MessageInputApiTestSuite) SetupTest() {
	s.manager = rabbitmq.
		NewRabbitMQManager(context.Background())
	s.manager.
		WithOptions(nil).
		WithGraceTimeOut(graceTimeOut).
		WithConnectionTimeOut(connTimeOut).
		WithHearBeat(heartBeat).
		WithKeepAlive(true)
	s.manager.
		UseConnection(s.Host, s.Port, s.User, s.Password)
	s.manager.
		Connect()

	s.topology = rabbitmq.
		NewRabbitMQTopology("topology", context.Background())
	s.topology.
		UseManager(s.manager)
	s.topology.
		WithOptions(nil).
		WithGraceTimeOut(graceTimeOut).
		WithDeclareTimeOut(declareTimeOut).
		WithQueuesPurged(true)	
	s.topology.
		Topic("campaign_messages").Queue("input_messages").Bind("input_messages")

	s.consumer = rabbitmq.
		NewConsumer("consumer", context.Background())
	s.consumer.
		UseManager(s.manager)
	s.consumer.
		WithOptions(nil).
		WithGraceTimeOut(graceTimeOut).
		WithQueue("input_messages")
}

func (s *MessageInputApiTestSuite) TearDownTest() {
	s.manager.Disconnect()
}

func (s *MessageInputApiTestSuite) WaitForMessageQueueReady(wait time.Duration, tick time.Duration) {
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
			time.Sleep(declareTimeOut)
			return
		}
	}
}

func (s *MessageInputApiTestSuite) TestCampaignMessageInputApi() {
	now := time.Now().Format("2006-01-02 15:04:05")
	// Wait for the message queue ready
	s.WaitForMessageQueueReady(10 * time.Second, graceTimeOut)
	// Send HTTP API Request
	url := "http://localhost:8002/campaign/test_campaign/message"
	jsonData := fmt.Sprintf(`{ "message": "%v" }`, now)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer([]byte(jsonData)))
	if err != nil {
		log.Println(err)
		s.FailNow("message input api request failure")
	}
	s.Require().Equal(resp.StatusCode, http.StatusCreated, "api must return 201 status code")
	// Validate the message has been pushed successfully 
	done := make(chan bool, 1)
	s.consumer.Consume(done, func(mssg string) mq.ConsumerAction {
		defer func() {
			done <- true
		}()

		var message model.InputMessage
		json.Unmarshal([]byte(mssg), &message)

		s.Require().Equal(message.Campaign, "test_campaign", "message's campaign must be correct")
		s.Require().Equal(message.Content, now, "messages's content must be correct")
		s.Require().False(message.IsRelayed, "messages's 'relay flag' must not be false")

		return mq.ConsumerAccept
	})
}
