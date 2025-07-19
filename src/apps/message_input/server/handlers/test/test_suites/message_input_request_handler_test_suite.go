package test_suites

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"duolingo/apps/message_input/server"
	container "duolingo/libraries/dependencies_container"
	ps "duolingo/libraries/message_queue/pub_sub"
	"duolingo/models"

	"github.com/stretchr/testify/suite"
)

type MessageInputRequestTestSuite struct {
	suite.Suite

	ctx                context.Context
	messageInputServer *server.MessageInputApiServer
	msgInpSubscriber   ps.Subscriber
}

func NewMessageInputRequestTestSuite(
	ctx context.Context,
	messageInputServer *server.MessageInputApiServer,
) *MessageInputRequestTestSuite {
	return &MessageInputRequestTestSuite{
		ctx:                ctx,
		messageInputServer: messageInputServer,
		msgInpSubscriber:   container.MustResolveAlias[ps.Subscriber]("message_input_subscriber"),
	}
}

func (s *MessageInputRequestTestSuite) SetupTest() {
	go s.messageInputServer.Serve(s.ctx)
}

func (s *MessageInputRequestTestSuite) TearDownTest() {
	s.messageInputServer.Shutdown()
}

func (s *MessageInputRequestTestSuite) Test_HandleMessageInputRequest() {
	response, requestErr := s.makeRequest("testcampaign", "testtitle", "testbody")
	if !s.Assert().NoError(requestErr) {
		return
	}

	done := make(chan bool, 1)
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		select {
		case <-ctx.Done():
			s.FailNow("did not receive input message before timeout")
			return
		case <-done:
			return
		}
	}()
	go func() {
		defer wg.Done()
		consumeErr := s.msgInpSubscriber.ListeningMainTopic(ctx, func(ctx context.Context, msg string) {
			defer func() { done <- true }()
			s.Assert().NotPanics(func() {
				s.Assert().Equal(response.Data, models.MessageInputDecode([]byte(msg)))
			})
		})
		s.Assert().NoError(consumeErr)
	}()
	wg.Wait()
}

func (s *MessageInputRequestTestSuite) Test_Validation_Require_Params_Missing() {
	response, requestErr := s.makeRequest("testcampaign", "", "")
	if !s.Assert().NoError(requestErr) {
		return
	}
	s.Assert().False(response.Success)
	s.Assert().NotEmpty(response.Errors)
	s.Assert().Equal("message title must not empty", response.Errors["title"])
	s.Assert().Equal("message body must not empty", response.Errors["body"])
}

func (s *MessageInputRequestTestSuite) makeRequest(
	campaign string,
	title string,
	body string,
) (*MessageInputResponse, error) {
	data, _ := json.Marshal(map[string]string{
		"title": title,
		"body":  body,
	})
	endpoint := fmt.Sprintf(
		"http://%v/api/v1/campaigns/%v/message-input",
		s.messageInputServer.Addr(),
		campaign,
	)
	response, requestErr := http.Post(
		endpoint,
		"application/json",
		bytes.NewBuffer(data),
	)
	if requestErr != nil {
		return nil, requestErr
	}

	responseBody := new(MessageInputResponse)
	rawBody, _ := io.ReadAll(response.Body)
	response.Body.Close()
	json.Unmarshal(rawBody, responseBody)

	return responseBody, nil
}

type MessageInputResponse struct {
	Status  int                  `json:"status"`
	Success bool                 `json:"success"`
	Errors  map[string]string    `json:"errors"`
	Data    *models.MessageInput `json:"data"`
}
