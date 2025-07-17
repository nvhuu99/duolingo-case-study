package test_suites

import (
	"context"
	"fmt"
	"sync"
	"time"

	ps "duolingo/libraries/message_queue/pub_sub"

	"github.com/stretchr/testify/suite"
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

func (s *PubSubTestSuite) Test_Notify_And_Listening() {
	msgTotal := 0
	msgCount1 := 0
	msgCount2 := 0
	tp1 := "s1_topic"
	tp2 := "s2_topic"

	// start timeout
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()
	wg := new(sync.WaitGroup)
	wg.Add(3)
	go func() {
		defer wg.Done()
		<-ctx.Done()
		if msgCount1 != 2 || msgCount2 != 2 || msgTotal != 4 {
			s.FailNow(fmt.Sprintf(
				"expected %v messages, received %v messages", 4, msgTotal,
			))
		}
	}()

	// subscribe and listening
	subErr1 := s.firstSubscriber.Subscribe(tp1)
	subErr2 := s.secondSubscriber.Subscribe(tp2)
	if !s.Assert().NoError(subErr1) || !s.Assert().NoError(subErr2) {
		return
	}
	defer s.firstSubscriber.UnSubscribe(tp1)
	defer s.secondSubscriber.UnSubscribe(tp2)
	go func() {
		defer wg.Done()
		expectMsg := "s1_m1"
		err := s.firstSubscriber.Listening(ctx, tp1, func(msg string) {
			if s.Assert().Equal(expectMsg, msg) {
				msgCount1++
				msgTotal++
				if msg == "s1_m1" {
					expectMsg = "s1_m2"
				}
			}
			if msgTotal == 4 {
				cancel()
			}
		})
		s.Assert().NoError(err)
	}()
	go func() {
		defer wg.Done()
		expectMsg := "s2_m1"
		err := s.secondSubscriber.Listening(ctx, tp2, func(msg string) {
			if s.Assert().Equal(expectMsg, msg) {
				msgCount2++
				msgTotal++
				if msg == "s2_m1" {
					expectMsg = "s2_m2"
				}
			}
			if msgTotal == 4 {
				cancel()
			}
		})
		s.Assert().NoError(err)
	}()

	// after subscribed, publish messages
	pubErr1 := s.publisher.Notify(tp1, "s1_m1")
	pubErr2 := s.publisher.Notify(tp1, "s1_m2")
	pubErr3 := s.publisher.Notify(tp2, "s2_m1")
	pubErr4 := s.publisher.Notify(tp2, "s2_m2")
	if !s.Assert().NoError(pubErr1) || !s.Assert().NoError(pubErr2) ||
		!s.Assert().NoError(pubErr3) || !s.Assert().NoError(pubErr4) {
		return
	}

	wg.Wait()
}

func (s *PubSubTestSuite) Test_MainTopic_NotSet() {
	declareErr := s.publisher.DeclareMainTopic()
	subErr1 := s.firstSubscriber.SubscribeMainTopic()
	subErr2 := s.secondSubscriber.SubscribeMainTopic()
	s.Assert().Equal(ps.ErrPublisherMainTopicNotSet, declareErr)
	s.Assert().Equal(ps.ErrSubscriberMainTopicNotSet, subErr1)
	s.Assert().Equal(ps.ErrSubscriberMainTopicNotSet, subErr2)
}

func (s *PubSubTestSuite) Test_MainTopic_Notify_And_Listening() {
	mainTopic := "t1"
	firstSubReceived := false
	secondSubReceived := false
	totalMsg := 0

	// start timeout
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()
	wg := new(sync.WaitGroup)
	wg.Add(3)
	go func() {
		defer wg.Done()
		<-ctx.Done()
		if !firstSubReceived || !secondSubReceived {
			s.FailNow("main topic messages are not delivered")
		}
		if totalMsg != 2 {
			s.FailNow("expected 2 messages, received %v messages", totalMsg)
		}
	}()

	// set main topic and subscribe
	s.publisher.SetMainTopic(mainTopic)
	s.firstSubscriber.SetMainTopic(mainTopic)
	s.secondSubscriber.SetMainTopic(mainTopic)
	declareErr := s.publisher.DeclareMainTopic()
	subErr1 := s.firstSubscriber.SubscribeMainTopic()
	subErr2 := s.secondSubscriber.SubscribeMainTopic()
	if !s.Assert().NoError(declareErr) ||
		!s.Assert().NoError(subErr1) ||
		!s.Assert().NoError(subErr2) {
		return
	}
	go func() {
		defer wg.Done()
		err := s.firstSubscriber.ListeningMainTopic(ctx, func(msg string) {
			if s.Assert().Equal("test message", msg) {
				firstSubReceived = true
			}
			totalMsg++
			if totalMsg == 2 {
				cancel()
			}
		})
		s.Assert().NoError(err)
	}()
	go func() {
		defer wg.Done()
		err := s.secondSubscriber.ListeningMainTopic(ctx, func(msg string) {
			if s.Assert().Equal("test message", msg) {
				secondSubReceived = true
			}
			totalMsg++
			if totalMsg == 2 {
				cancel()
			}
		})
		s.Assert().NoError(err)
	}()

	// after subscribe main topic, publish message
	publishErr := s.publisher.NotifyMainTopic("test message")
	if !s.Assert().NoError(publishErr) {
		return
	}

	wg.Wait()

	removeErr := s.publisher.RemoveMainTopic()
	unSubErr1 := s.firstSubscriber.UnSubscribeMainTopic()
	unSubErr2 := s.secondSubscriber.UnSubscribeMainTopic()
	s.Assert().NoError(removeErr)
	s.Assert().NoError(unSubErr1)
	s.Assert().NoError(unSubErr2)
}
