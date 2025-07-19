package test_suites

import (
	"context"
	"slices"
	"sync"
	"time"

	"duolingo/apps/push_sender/server"
	"duolingo/apps/push_sender/server/test/fakes"
	"duolingo/libraries/config_reader"
	container "duolingo/libraries/dependencies_container"
	tq "duolingo/libraries/message_queue/task_queue"
	"duolingo/libraries/push_notification"
	"duolingo/models"
	"duolingo/test/fixtures/data"

	"github.com/stretchr/testify/suite"
)

type SenderTestSuite struct {
	suite.Suite

	config   config_reader.ConfigReader
	producer tq.TaskProducer
}

func NewSenderTestSuite() *SenderTestSuite {
	return &SenderTestSuite{
		config:   container.MustResolve[config_reader.ConfigReader](),
		producer: container.MustResolveAlias[tq.TaskProducer]("push_notifications_producer"),
	}
}

func (s *SenderTestSuite) Test_Sender_BufferLimit() {
	pushService := container.MustResolve[push_notification.PushService]()
	fakeService, ok := pushService.(*fakes.FakePushService)
	if !ok {
		panic("canot resolve fake push service")
	}
	sender := server.NewSender()

	var firstMessage *fakes.FakeMessage
	var secMessage *fakes.FakeMessage
	var totalToken = 2 * len(data.TestDevices) // nums messages * nums devices
	var totalFlush = totalToken / s.config.GetInt("push_sender", "buffer_limit")
	var flushTokenCount = 0
	var flushCount = 0

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	wg := new(sync.WaitGroup)
	wg.Add(3)
	go func() {
		defer wg.Done()
		sender.Start(ctx)
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
				if flushCount == totalFlush {
					cancel()
				}
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
		s.producer.Push(string(noti1.Encode()))
		s.producer.Push(string(noti2.Encode()))
	}()
	wg.Wait()
}
