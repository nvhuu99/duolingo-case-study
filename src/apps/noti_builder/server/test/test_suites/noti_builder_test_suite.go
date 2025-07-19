package test_suites

import (
	"context"
	"sync"
	"time"

	"duolingo/apps/noti_builder/server"
	container "duolingo/libraries/dependencies_container"
	ps "duolingo/libraries/message_queue/pub_sub"
	tq "duolingo/libraries/message_queue/task_queue"
	dist "duolingo/libraries/work_distributor"
	"duolingo/models"
	usr_repo "duolingo/repositories/user_repository/external"
	"duolingo/test/fixtures/data"

	"github.com/stretchr/testify/suite"
)

type NotiBuilderTestSuite struct {
	suite.Suite
	usrRepo          usr_repo.UserRepository
	inputPublisher   ps.Publisher
	pushNotiConsumer tq.TaskConsumer
	distributor      *dist.WorkDistributor
	builder          *server.NotiBuilder
}

func NewNotiBuilderTestSuite(builder *server.NotiBuilder) *NotiBuilderTestSuite {
	return &NotiBuilderTestSuite{
		usrRepo:          container.MustResolve[usr_repo.UserRepository](),
		inputPublisher:   container.MustResolveAlias[ps.Publisher]("message_input_publisher"),
		pushNotiConsumer: container.MustResolveAlias[tq.TaskConsumer]("push_notifications_consumer"),
		distributor:      container.MustResolve[*dist.WorkDistributor](),
		builder:          builder,
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
	totalBatches := totalDevices / int(s.distributor.GetDistributionSize())
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
		s.builder.Start(ctx)
	}()
	go func() {
		defer wg.Done()
		err := s.pushNotiConsumer.Consuming(ctx, func(ctx context.Context, str string) {
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
		})
		s.Assert().NoError(err)
	}()
	wg.Wait()
}
