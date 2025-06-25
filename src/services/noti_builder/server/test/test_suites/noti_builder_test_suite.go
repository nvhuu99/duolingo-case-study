package test_suites

import (
	"context"
	"duolingo/libraries/pub_sub"
	"duolingo/models"
	"duolingo/repositories/user_repository/external"
	"duolingo/services/noti_builder/server"
	cnst "duolingo/test/fixtures/constants"
	"duolingo/test/fixtures/data"
	"sync"
	"time"

	"github.com/stretchr/testify/suite"
)

type NotiBuilderTestSuite struct {
	suite.Suite
	usrRepo        external.UserRepository
	inputPublisher pub_sub.Publisher
	notiSubscriber pub_sub.Subscriber
	builder        *server.NotiBuilder
}

func NewNotiBuilderTestSuite(
	usrRepo external.UserRepository,
	inputPublisher pub_sub.Publisher,
	notiSubscriber pub_sub.Subscriber,
	notiBuilder *server.NotiBuilder,
) *NotiBuilderTestSuite {
	return &NotiBuilderTestSuite{
		usrRepo:        usrRepo,
		inputPublisher: inputPublisher,
		notiSubscriber: notiSubscriber,
		builder:        notiBuilder,
	}
}

func (s *NotiBuilderTestSuite) SetupTest() {
	s.usrRepo.DeleteUsersByIds(data.TestUserIds)
	s.usrRepo.InsertManyUsers(data.TestUsers)
}

func (s *NotiBuilderTestSuite) TearDownTest() {
	s.usrRepo.DeleteUsersByIds(data.TestUserIds)
}

func (s *NotiBuilderTestSuite) Test_NotiBuilder() {
	input1 := models.NewMessageInput(data.TestCampaignPrimary, "title 1", "body 1")
	input2 := models.NewMessageInput(data.TestCampaignPrimary, "title 2", "body 2")
	s.inputPublisher.NotifyMainTopic(string(input1.Encode()))
	s.inputPublisher.NotifyMainTopic(string(input2.Encode()))

	totalDevices := 2 * len(data.TestDevices)
	totalBatches := totalDevices / int(cnst.DistributionSize)
	countDevices := 0
	countBatches := 0

	done := make(chan bool, 1)
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	var wg sync.WaitGroup
	wg.Add(3)
	go func() {
		defer wg.Done()
		select {
		case <-ctx.Done():
			s.FailNow("did not receive all batches before timeout")
			return
		case <-done:
			return
		}
	}()
	go func() {
		defer wg.Done()
		err := s.builder.Start(ctx)
		s.Assert().NoError(err)
	}()
	go func() {
		defer wg.Done()
		err := s.notiSubscriber.ConsumingMainTopic(ctx, func(str string) pub_sub.ConsumeAction {
			s.Assert().NotPanics(func() {
				pushNoti := models.PushNotiMessageDecode([]byte(str))
				if pushNoti.MessageInput == nil {
					return
				}
				devices := pushNoti.GetTargetTokens(data.TestPlatforms)
				countDevices += len(devices)
				countBatches++

				s.Assert().True(
					*input1 == *pushNoti.MessageInput ||
						*input2 == *pushNoti.MessageInput,
				)
				s.Assert().NotEmpty(devices)

				if countBatches == totalBatches && countDevices == totalDevices {
					done <- true
				}
			})
			return pub_sub.ActionAccept
		})
		s.Assert().NoError(err)
	}()
	wg.Wait()
}
