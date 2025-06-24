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

	publisher        ps.Publisher
	firstSubscriber  ps.Subscriber
	secondSubscriber ps.Subscriber
	subscribers      []ps.Subscriber
}

func NewPubSubTestSuite(
	publisher ps.Publisher,
	firstSubscriber ps.Subscriber,
	secondSubscriber ps.Subscriber,
) *PubSubTestSuite {
	return &PubSubTestSuite{
		publisher:        publisher,
		firstSubscriber:  firstSubscriber,
		secondSubscriber: secondSubscriber,
		subscribers:      []ps.Subscriber{firstSubscriber, secondSubscriber},
	}
}

func (s *PubSubTestSuite) SetupSuite() {
	if topicErr := s.publisher.DeclareTopic(TestTopic); topicErr != nil {
		s.FailNow("failed to declare the test topic")
	}
	for _, sub := range s.subscribers {
		if subscribeErr := sub.Subscribe(TestTopic); subscribeErr != nil {
			s.FailNow("fail to subscribe to the test topic")
		}
	}
}

func (s *PubSubTestSuite) TearDownSuite() {
	for _, sub := range s.subscribers {
		sub.UnSubscribe(TestTopic)
	}
	s.publisher.RemoveTopic(TestTopic)
}

func (s *PubSubTestSuite) Test_Notify_And_Consuming() {
	firstSubscriberTopic := "first_subscriber_test_topic"
	subErr1 := s.firstSubscriber.Subscribe(firstSubscriberTopic)
	defer s.firstSubscriber.UnSubscribe(firstSubscriberTopic)

	secondSubscriberTopic := "second_subscriber_test_topic"
	subErr2 := s.secondSubscriber.Subscribe(secondSubscriberTopic)
	defer s.secondSubscriber.UnSubscribe(secondSubscriberTopic)

	publishErr1 := s.publisher.Notify(firstSubscriberTopic, "first_subscriber_test_message_1")
	publishErr2 := s.publisher.Notify(firstSubscriberTopic, "first_subscriber_test_message_2")
	if !s.Assert().NoError(subErr1) || !s.Assert().NoError(subErr2) {
		return
	}

	publishErr3 := s.publisher.Notify(secondSubscriberTopic, "second_subscriber_test_message_1")
	publishErr4 := s.publisher.Notify(secondSubscriberTopic, "second_subscriber_test_message_2")
	if !s.Assert().NoError(publishErr1) || !s.Assert().NoError(publishErr2) ||
		!s.Assert().NoError(publishErr3) || !s.Assert().NoError(publishErr4) {
		return
	}

	firstSubMsgCount := 0
	secondSubMsgCount := 0
	allMsgCount := 0
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	wg := new(sync.WaitGroup)
	wg.Add(3)
	go func() {
		defer wg.Done()
		<-ctx.Done()
		if firstSubMsgCount != 2 || secondSubMsgCount != 2 || allMsgCount != 4 {
			s.FailNow("messages delivered mismatch expected")
		}
	}()
	go func() {
		defer wg.Done()
		expectMsg := "first_subscriber_test_message_1"
		err := s.firstSubscriber.Consuming(ctx, firstSubscriberTopic, func(msg string) ps.ConsumeAction {
			if s.Assert().Equal(expectMsg, msg) {
				firstSubMsgCount++
				allMsgCount++
				if msg == "first_subscriber_test_message_1" {
					expectMsg = "first_subscriber_test_message_2"
				}
			}
			return ps.ActionAccept
		})
		s.Assert().NoError(err)
	}()
	go func() {
		defer wg.Done()
		expectMsg := "second_subscriber_test_message_1"
		err := s.secondSubscriber.Consuming(ctx, secondSubscriberTopic, func(msg string) ps.ConsumeAction {
			if s.Assert().Equal(expectMsg, msg) {
				secondSubMsgCount++
				allMsgCount++
				if msg == "second_subscriber_test_message_1" {
					expectMsg = "second_subscriber_test_message_2"
				}
			}
			return ps.ActionAccept
		})
		s.Assert().NoError(err)
	}()

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
