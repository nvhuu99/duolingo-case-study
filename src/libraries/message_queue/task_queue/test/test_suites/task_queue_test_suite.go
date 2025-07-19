package test_suites

import (
	"context"
	tq "duolingo/libraries/message_queue/task_queue"
	"slices"
	"sync"
	"time"

	"github.com/stretchr/testify/suite"
)

type TaskQueueTestSuite struct {
	suite.Suite
	queue         tq.TaskQueue
	producer      tq.TaskProducer
	firstConsumer tq.TaskConsumer
	secConsumer   tq.TaskConsumer
}

func NewTaskQueueTestSuite(
	queue tq.TaskQueue,
	producer tq.TaskProducer,
	firstConsumer tq.TaskConsumer,
	secConsumer tq.TaskConsumer,
) *TaskQueueTestSuite {
	return &TaskQueueTestSuite{
		queue:         queue,
		producer:      producer,
		firstConsumer: firstConsumer,
		secConsumer:   secConsumer,
	}
}

func (s *TaskQueueTestSuite) Test_Produce_And_Consume() {
	s.queue.SetQueue("test_tq")
	s.producer.SetQueue("test_tq")
	s.firstConsumer.SetQueue("test_tq")
	s.secConsumer.SetQueue("test_tq")
	defer s.queue.Remove()

	declareErr := s.queue.Declare()
	if !s.Assert().NoError(declareErr) {
		return
	}

	taskCount1 := 0
	taskCount2 := 0
	totalTasks := 0
	tasks := []string{"t1", "t2", "t3", "t4"}

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()
	wg := new(sync.WaitGroup)
	wg.Add(4)
	go func() {
		defer wg.Done()
		<-ctx.Done()
		s.Assert().Equal(2, taskCount1, "first consumer receive 2 tasks")
		s.Assert().Equal(2, taskCount2, "second consumer receive 2 tasks")
		s.Assert().Equal(4, totalTasks, "all tasks received")
	}()
	go func() {
		defer wg.Done()
		t := 0
		for {
			select {
			case <-ctx.Done():
				return
			default:
			}
			if t == len(tasks) {
				return
			}
			err := s.producer.Push(tasks[t])
			if !s.Assert().NoError(err) {
				cancel()
			}
			t++
		}
	}()
	go func() {
		defer wg.Done()
		err := s.firstConsumer.Consuming(ctx, func(ctx context.Context, t string) {
			if s.Assert().True(slices.Contains(tasks, t)) {
				taskCount1++
				totalTasks++
			}
			if totalTasks == 4 {
				cancel()
			}
		})
		s.Assert().NoError(err)
	}()
	go func() {
		defer wg.Done()
		err := s.secConsumer.Consuming(ctx, func(ctx context.Context, t string) {
			if s.Assert().True(slices.Contains(tasks, t)) {
				taskCount2++
				totalTasks++
			}
			if totalTasks == 4 {
				cancel()
			}
		})
		s.Assert().NoError(err)
	}()
	wg.Wait()

}
