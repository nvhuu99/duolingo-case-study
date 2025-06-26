package test_suites

import (
	"context"
	"duolingo/libraries/pub_sub"
	"duolingo/models"
	"duolingo/services/push_sender/server"
	"duolingo/services/push_sender/server/test/fakes"
	"duolingo/test/fixtures/data"
	"slices"
	"sync"
	"time"

	"github.com/stretchr/testify/suite"
)

type SenderTestSuite struct {
	suite.Suite

	notiPublisher  pub_sub.Publisher
	notiSubscriber pub_sub.Subscriber
}

func NewSenderTestSuite(
	notiPublisher pub_sub.Publisher,
	notiSubscriber pub_sub.Subscriber,
) *SenderTestSuite {
	return &SenderTestSuite{
		notiPublisher:  notiPublisher,
		notiSubscriber: notiSubscriber,
	}
}

func (s *SenderTestSuite) Test_Sender_BufferLimit() {
	limit := 5
	interval := 10 * time.Millisecond
	fakeService := fakes.NewFakePushService()
	sender := server.NewSender(
		s.notiSubscriber,
		fakeService,
		data.TestPlatforms,
		limit,
		interval,
	)

	var firstMessage *fakes.FakeMessage
	var secMessage *fakes.FakeMessage
	var totalToken = 2 * len(data.TestDevices) // nums messages * nums devices
	var totalFlush = totalToken / limit
	var flushTokenCount = 0
	var flushCount = 0

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	wg := new(sync.WaitGroup)
	wg.Add(3)
	go func() {
		defer wg.Done()
		err := sender.Start(ctx)
		s.Assert().NoError(err)
	}()
	go func() {
		defer wg.Done()
		msgChan := fakeService.GetMesgChan()
		for {
			select {
			case <-ctx.Done():
				s.Assert().Equal(totalToken, flushTokenCount)
				s.Assert().Equal(totalFlush, flushCount)
				if s.Assert().NotNil(firstMessage) {
					s.Assert().Equal("title 1", firstMessage.Title)
					s.Assert().Equal("body 1", firstMessage.Body)
					for i := range data.TestDevices {
						s.Assert().True(slices.Contains(
							firstMessage.Tokens, data.TestDevices[i].Token))
					}
					for i := range data.TestPlatforms {
						s.Assert().True(slices.Contains(
							firstMessage.Platforms, data.TestPlatforms[i]))
					}
				}
				if s.Assert().NotNil(secMessage) {
					s.Assert().Equal("title 2", secMessage.Title)
					s.Assert().Equal("body 2", secMessage.Body)
					for i := range data.TestDevices {
						s.Assert().True(slices.Contains(
							secMessage.Tokens, data.TestDevices[i].Token))
					}
					for i := range data.TestPlatforms {
						s.Assert().True(slices.Contains(
							secMessage.Platforms, data.TestPlatforms[i]))
					}
				}
				return
			case msg := <-msgChan:
				if msg.Title == "title 1" {
					if firstMessage == nil {
						firstMessage = &fakes.FakeMessage{
							Title:     msg.Title,
							Body:      msg.Body,
							Platforms: msg.Platforms,
						}
					}
					firstMessage.Tokens = append(firstMessage.Tokens, msg.Tokens...)
				}
				if msg.Title == "title 2" {
					if secMessage == nil {
						secMessage = &fakes.FakeMessage{
							Title:     msg.Title,
							Body:      msg.Body,
							Platforms: msg.Platforms,
						}
					}
					secMessage.Tokens = append(secMessage.Tokens, msg.Tokens...)
				}
				flushCount++
				flushTokenCount += len(msg.Tokens)
			}
		}
	}()
	go func() {
		defer wg.Done()
		noti1 := models.NewPushNotiMessage(
			models.NewMessageInput(data.TestCampaignPrimary, "title 1", "body 1"),
			data.TestDevices,
		)
		noti2 := models.NewPushNotiMessage(
			models.NewMessageInput(data.TestCampaignPrimary, "title 2", "body 2"),
			data.TestDevices,
		)
		s.notiPublisher.NotifyMainTopic(string(noti1.Encode()))
		s.notiPublisher.NotifyMainTopic(string(noti2.Encode()))
	}()
	wg.Wait()
}
