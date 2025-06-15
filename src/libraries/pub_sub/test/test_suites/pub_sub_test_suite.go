package test_suites

import (
	"context"
	"sync"
	"time"

	ps "duolingo/libraries/pub_sub"

	"github.com/stretchr/testify/suite"
)

const (
	TestTopic = "TEST_ONLY_TOPIC"
)

type PubSubTestSuite struct {
	suite.Suite

	publisher   ps.Publisher
	subscribers []ps.Subscriber
}

func NewPubSubTestSuite(
	publisher ps.Publisher,
	subscribers []ps.Subscriber,
) *PubSubTestSuite {
	return &PubSubTestSuite{
		publisher:   publisher,
		subscribers: subscribers,
	}
}

func (s *PubSubTestSuite) SetupSuite() {
	for _, sub := range s.subscribers {
		s.publisher.AddSubscriber(TestTopic, sub)
	}
}

func (s *PubSubTestSuite) TeardownSuite() {
	s.T().Log("Teardown")
	for _, sub := range s.subscribers {
		s.publisher.RemoveSubscriber(sub)
	}
}

func (s *PubSubTestSuite) Test_Notify_And_Consuming() {
	publishErr := s.publisher.Notify(TestTopic, "test message")
	if !s.Assert().NoError(publishErr) {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	var wg sync.WaitGroup
	wg.Add(len(s.subscribers))
	for _, sub := range s.subscribers {
		go func() {
			defer wg.Done()
			consumeErr := sub.Consuming(ctx, TestTopic, func(msg string) ps.ConsumeAction {
				s.Assert().Equal("test message", msg)
				return ps.ActionAccept
			})
			s.Assert().NoError(consumeErr)
		}()
	}
	wg.Wait()
}

func (s *PubSubTestSuite) Test_Requeue_Message() {
	publishErr := s.publisher.Notify(TestTopic, "test message requeue")
	if !s.Assert().NoError(publishErr) {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	var wg sync.WaitGroup
	wg.Add(len(s.subscribers))
	for _, sub := range s.subscribers {
		go func() {
			defer wg.Done()
			receiveCount := 0
			isRequeued := false
			consumeErr := sub.Consuming(ctx, TestTopic, func(msg string) ps.ConsumeAction {
				if s.Assert().Equal("test message requeue", msg) {
					receiveCount++
				}
				if !isRequeued {
					isRequeued = true
					return ps.ActionRequeue
				}
				return ps.ActionAccept
			})
			s.Assert().NoError(consumeErr)
			s.Assert().Equal(2, receiveCount)
		}()
	}
	wg.Wait()
}

func (s *PubSubTestSuite) Test_Reject_Message() {
	publishErr := s.publisher.Notify(TestTopic, "test message reject")
	if !s.Assert().NoError(publishErr) {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	var wg sync.WaitGroup
	wg.Add(len(s.subscribers))
	for _, sub := range s.subscribers {
		go func() {
			defer wg.Done()
			receiveCount := 0
			consumeErr := sub.Consuming(ctx, TestTopic, func(msg string) ps.ConsumeAction {
				if s.Assert().Equal("test message reject", msg) {
					receiveCount++
				}
				return ps.ActionReject
			})
			s.Assert().NoError(consumeErr)
			s.Assert().Equal(1, receiveCount)
		}()
	}
	wg.Wait()
}
