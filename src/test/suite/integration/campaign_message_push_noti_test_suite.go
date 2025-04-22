// Warning: DO NOT run the test on production environment, data loss could occur.
// Only run it in test environments.
package test_suite

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"time"

	cnst "duolingo/constant"
	config "duolingo/lib/config_reader"
	lq "duolingo/lib/log/driver/reader/local"
	"duolingo/lib/message_queue/driver/rabbitmq"
	"duolingo/model"
	ldt "duolingo/model/log/detail"
	db "duolingo/repository/campaign_db"

	"github.com/stretchr/testify/suite"
)

var (
	graceTimeOut   = 200 * time.Millisecond
	connTimeOut    = 1 * time.Second
	declareTimeOut = 1 * time.Second
	heartBeat      = 1 * time.Second
)

// CampaignMessagePushNotiTestSuite defines the test suite for campaign message push notification
// covering message input, relay, build, and sending operations.
type CampaignMessagePushNotiTestSuite struct {
	suite.Suite
	ServiceDir   string
	ConfigReader config.ConfigReader
	Campaign     string

	manager      *rabbitmq.RabbitMQManager
	topology     *rabbitmq.RabbitMQTopology
	inputMessage *model.InputMessage
	traceId      string
	userCount    int
}

// SetupSuite initializes DB connections, RabbitMQ manager, and topology before tests
func (s *CampaignMessagePushNotiTestSuite) SetupSuite() {
	ctx := context.Background()
	repo := db.NewUserRepo(ctx, s.ConfigReader.Get("db.campaign.name", ""))
	repo.SetConnection(
		s.ConfigReader.Get("db.campaign.host", ""),
		s.ConfigReader.Get("db.campaign.port", ""),
		s.ConfigReader.Get("db.campaign.user", ""),
		s.ConfigReader.Get("db.campaign.password", ""),
	)

	s.userCount, _ = repo.CountUsers(s.Campaign)

	s.manager = rabbitmq.NewRabbitMQManager(ctx)
	s.manager.
		WithOptions(nil).
		WithGraceTimeOut(graceTimeOut).
		WithConnectionTimeOut(connTimeOut).
		WithHearBeat(heartBeat).
		WithKeepAlive(true)
	s.manager.
		UseConnection(
			s.ConfigReader.Get("mq.host", ""),
			s.ConfigReader.Get("mq.port", ""),
			s.ConfigReader.Get("mq.user", ""),
			s.ConfigReader.Get("mq.pwd", ""),
		)
	s.manager.
		Connect()

	s.topology = rabbitmq.
		NewRabbitMQTopology("campaign_messages_topology", ctx)
	s.topology.
		UseManager(s.manager)
	s.topology.
		WithOptions(nil).
		WithGraceTimeOut(graceTimeOut).
		WithDeclareTimeOut(declareTimeOut).
		WithQueuesPurged(true)
	s.topology.
		Topic("campaign_messages").Queue("input_messages").Bind("input_messages")
	s.topology.
		Topic("campaign_messages").Queue("push_noti_messages").Bind("push_noti_messages")
}

// TearDownSuite disconnects from RabbitMQ after tests
func (s *CampaignMessagePushNotiTestSuite) TearDownSuite() {
	s.manager.Disconnect()
}

// TestStep01MessageInputAPI tests the message input API endpoint and ensures it logs correctly
func (s *CampaignMessagePushNotiTestSuite) TestStep01MessageInputAPI() {
	s.waitForMessageQueueReady(10*time.Second, graceTimeOut)

	campaign := "superbowl"
	addr := s.ConfigReader.Get("input_message_api.server.address", "")
	api := fmt.Sprintf("http://%v/campaign/%v/message", addr, campaign)
	requestBody := `{ "title": "test_title", "content": "test_content" }`

	// Verify API status
	resp, err := http.Post(api, "application/json", bytes.NewBuffer([]byte(requestBody)))
	s.Require().NoError(err, "POST request to input API should succeed")
	s.Require().Equal(http.StatusCreated, resp.StatusCode, "Expected HTTP 201 Created from input API")

	// Extract the input message from the response
	// and save it for next the later test cases
	var responseBody struct {
		Data *model.InputMessage `json:"data"`
	}
	rawBody, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	json.Unmarshal(rawBody, &responseBody)
	s.inputMessage = responseBody.Data

	// Verify message properties
	s.Require().Equal("test_title", s.inputMessage.Title, "Returned title must match input")
	s.Require().Equal("test_content", s.inputMessage.Content, "Returned content must match input")
	s.Require().Equal(campaign, s.inputMessage.Campaign, "Returned campaign must match input")

	// Query input message request log
	logQuery := s.logQuery(cnst.SV_INP_MESG).
		Info().FilePrefix(cnst.SV_INP_MESG).
		ExpectJson().
		Filter(map[string]any{
			"context": map[string]any{
				"trace": map[string]any{
					"service_name":      cnst.SV_INP_MESG,
					"service_operation": cnst.INP_MESG_REQUEST,
				},
			},
			"data": map[string]any{
				"request": map[string]any{
					"response_body_data": map[string]any{
						"id":       s.inputMessage.MessageId,
						"title":    s.inputMessage.Title,
						"content":  s.inputMessage.Content,
						"campaign": s.inputMessage.Campaign,
					},
				},
			},
		})

	s.waitForServiceLogs(logQuery, 10*time.Second, graceTimeOut)

	found, err := logQuery.First()
	s.Require().NoError(err, "Error finding input message request log")
	s.Require().True(len(found) != 0, "Input message request log not found")

	detail := new(ldt.InputMessageRequest)
	data, _ := json.Marshal(found)
	err = json.Unmarshal(data, detail)
	s.Require().NoError(err, "Error converting to log detail")

	// Save trace id
	s.traceId = detail.LogContext.Trace.TraceId
}

// TestStep02RelayInputMessageForAllBuilders verifies that the input message is relayed to all builders
func (s *CampaignMessagePushNotiTestSuite) TestStep02RelayInputMessageForAllBuilders() {
	logQuery := s.logQuery(cnst.SV_NOTI_BUILDER).
		Info().FilePrefix(cnst.SV_NOTI_BUILDER).
		ExpectJson().
		Filter(map[string]any{
			"context": map[string]any{
				"trace": map[string]any{
					"trace_id":          s.traceId,
					"service_type":      cnst.ServiceTypes[cnst.SV_NOTI_BUILDER],
					"service_name":      cnst.SV_NOTI_BUILDER,
					"service_operation": cnst.RELAY_INP_MESG,
				},
			},
			"data": map[string]any{
				"relayed_message": map[string]any{
					"id": s.inputMessage.MessageId,
				},
				"relayed_count": s.ConfigReader.GetInt("noti_builder.server.num_of_builders", -1),
			},
		})

	s.waitForServiceLogs(logQuery, 10*time.Second, graceTimeOut)

	// Verify relay operation log
	numOfRelayedOperation, err := logQuery.Count()
	s.Require().NoError(err, "Log query for relayed input must not return error")
	s.Require().Equal(1, numOfRelayedOperation, "Exactly one relay operation expected in logs")
}

// TestStep03BuildPushNotiForAllUsers verifies the push notifications were built successfully
func (s *CampaignMessagePushNotiTestSuite) TestStep03BuildPushNotiForAllUsers() {
	logQuery := s.logQuery(cnst.SV_NOTI_BUILDER).
		Info().FilePrefix(cnst.SV_NOTI_BUILDER).
		ExpectJson().
		Filter(map[string]any{
			"context": map[string]any{
				"trace": map[string]any{
					"trace_id":          s.traceId,
					"service_type":      cnst.ServiceTypes[cnst.SV_NOTI_BUILDER],
					"service_name":      cnst.SV_NOTI_BUILDER,
					"service_operation": cnst.BUILD_PUSH_NOTI_MESG,
				},
			},
			"data": map[string]any{
				"message": map[string]any{
					"id": s.inputMessage.MessageId,
				},
			},
		})

	s.waitForServiceLogs(logQuery, 10*time.Second, graceTimeOut)

	allSuccess := true
	totalMessages := 0
	totalOperations := 0

	err := logQuery.Each(func(item map[string]any) lq.LoopAction {
		var detail ldt.BuildNotification
		data, _ := json.Marshal(item)
		err := json.Unmarshal(data, &detail)
		if err != nil {
			s.FailNow("Error deserializeing build notification log")
			return lq.LoopCancel
		}

		totalOperations++

		if len(detail.LogData.Assignments) == 0 {
			return lq.LoopContinue
		}
		for _, asgm := range detail.LogData.Assignments {
			totalMessages += asgm.End - asgm.Start + 1
		}

		return lq.LoopContinue
	})

	// Verify push notification messages build operations
	s.Require().NoError(err, "Error iterating build notification logs")
	s.Require().True(allSuccess, "Some build notification operations failed")
	s.Require().Equal(s.userCount, totalMessages, "User count and push notification message must match")
	s.Require().Equal(
		s.ConfigReader.GetInt("noti_builder.server.num_of_builders", -1),
		totalOperations,
		"Number of build operations and number of builders must match",
	)
}

// // TestStep04SendPushNotiForAllUsers ensures push notifications were sent successfully to all users
func (s *CampaignMessagePushNotiTestSuite) TestStep04SendPushNotiForAllUsers() {
	logQuery := s.logQuery(cnst.SV_PUSH_SENDER).
		Info().FilePrefix(cnst.SV_PUSH_SENDER).
		ExpectJson().
		Filter(map[string]any{
			"context": map[string]any{
				"trace": map[string]any{
					"trace_id":          s.traceId,
					"service_type":      cnst.ServiceTypes[cnst.SV_PUSH_SENDER],
					"service_name":      cnst.SV_PUSH_SENDER,
					"service_operation": cnst.SEND_PUSH_NOTI,
				},
			},
			"data": map[string]any{
				"message": map[string]any{
					"id": s.inputMessage.MessageId,
				},
			},
		})

	s.waitForServiceLogs(logQuery, 10*time.Second, graceTimeOut)

	allSuccess := true
	successCount := 0
	failureCount := 0
	err := logQuery.Each(func(item map[string]any) lq.LoopAction {
		var detail ldt.SendPushNotification
		data, _ := json.Marshal(item)
		err := json.Unmarshal(data, &detail)
		if err != nil {
			s.FailNow("Fail to deserializing send push notification log")
			return lq.LoopCancel
		}

		if !detail.LogData.Success {
			allSuccess = false
			return lq.LoopCancel
		}
		successCount += detail.LogData.SuccessCount
		failureCount += detail.LogData.FailureCount

		return lq.LoopContinue
	})

	// Verify push notification send operations
	s.Require().NoError(err, "Error iterating push notification logs")
	s.Require().True(allSuccess, "Some push notification deliveries failed")
	s.Require().Equal(s.userCount, successCount+failureCount, "Mismatch between user count and total tokens processed")
}

// waitForMessageQueueReady waits for RabbitMQ manager and topology to become ready
func (s *CampaignMessagePushNotiTestSuite) waitForMessageQueueReady(wait time.Duration, tick time.Duration) {
	timeOut := time.After(wait)
	for {
		select {
		case <-timeOut:
			return
		default:
			if !s.manager.IsReady() || !s.topology.IsReady() {
				time.Sleep(tick)
				continue
			}
			return
		}
	}
}

// waitForMessageQueueReady waits for RabbitMQ manager and topology to become ready
func (s *CampaignMessagePushNotiTestSuite) waitForServiceLogs(query *lq.LocalReader, wait time.Duration, tick time.Duration) {
	timeOut := time.After(wait)
	for {
		select {
		case <-timeOut:
			return
		default:
			if found, _ := query.Any(); found {
				time.Sleep(time.Second) // Gracefully wait services operations to be completed
				return
			}
			time.Sleep(tick)
		}
	}
}

// logQuery returns a local log reader for the given service for today
func (s *CampaignMessagePushNotiTestSuite) logQuery(service string) *lq.LocalReader {
	now := time.Now()
	from := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	to := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 0, 0, now.Location())
	dir := filepath.Join(
		s.ServiceDir,
		service,
		"storage", "log", "service",
		cnst.ServiceTypes[service],
	)

	return lq.LogQuery(dir, from, to)
}
