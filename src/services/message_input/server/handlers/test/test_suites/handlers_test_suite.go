package test_suites

import (
	"context"
	"duolingo/libraries/pub_sub"
	"duolingo/models"
	"duolingo/services/message_input/server/handlers"
	"duolingo/services/message_input/server/requests"
	"sync"
	"time"

	"github.com/stretchr/testify/suite"
)

type HandlersTestSuite struct {
	suite.Suite

	inputPublisher  pub_sub.Publisher
	inputSubscriber pub_sub.Subscriber
}

func NewHandlersTestSuite(
	inputPublisher pub_sub.Publisher,
	inputSubscriber pub_sub.Subscriber,
) *HandlersTestSuite {
	return &HandlersTestSuite{
		inputPublisher:  inputPublisher,
		inputSubscriber: inputSubscriber,
	}
}

func (s *HandlersTestSuite) Test_HandleMessageInputRequest() {
	request, _ := requests.NewMessageInputRequest(
		"superbowl",
		"test title",
		"test body",
	)
	handler := handlers.NewMessageInputRequestHandler(s.inputPublisher)
	input, handleErr := handler.Handle(request)

	if !s.Assert().NoError(handleErr) {
		return
	}

	done := make(chan bool, 1)
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
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
		consumeErr := s.inputSubscriber.ConsumingMainTopic(ctx, func(msg string) pub_sub.ConsumeAction {
			defer func() { done <- true }()

			decodedInp := new(models.MessageInput)
			s.Assert().NotPanics(func() {
				decodedInp = models.MessageInputDecode([]byte(msg))
			})
			s.Assert().Equal("superbowl", decodedInp.Campaign)
			s.Assert().Equal("test title", decodedInp.Title)
			s.Assert().Equal("test body", decodedInp.Body)
			s.Assert().Equal(input.Id, decodedInp.Id)

			return pub_sub.ActionAccept
		})
		s.Assert().NoError(consumeErr)
	}()
	wg.Wait()
}
