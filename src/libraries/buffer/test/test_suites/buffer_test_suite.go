package test_suites

import (
	"context"
	"duolingo/libraries/buffer"
	"strings"
	"sync"
	"time"

	"github.com/stretchr/testify/suite"
)

type BufferTestSuite struct {
	suite.Suite
}

func (s *BufferTestSuite) Test_Buffer_Limit() {
	done := make(chan bool, 1)
	buff := buffer.NewBuffer[string]()
	buff.SetLimit(3).
		SetInterval(100*time.Second). // this amount ensure the flush trigger by limit
		SetConsumeFunc(true, func(items []string) {
			defer buff.Stop()
			defer func() { done <- true }()
			if s.Assert().Equal(len(items), 3) {
				for i := range items {
					s.Assert().True(strings.HasPrefix(items[i], "test_item_"))
				}
			}
		}).
		Start(context.Background())

	timeout := time.After(10 * time.Millisecond)
	wg := new(sync.WaitGroup)
	wg.Add(1)
	go func() {
		defer wg.Done()
		select {
		case <-done:
			return
		case <-timeout:
			s.FailNow("buffer should flush before timeout")
		}
	}()

	buff.Write("test_item_1")
	buff.Write("test_item_2")
	buff.Write("test_item_3")
	buff.Write("test_item_4") // 3 items limit hit, should trigger flushing

	wg.Wait()
}

func (s *BufferTestSuite) Test_Buffer_Flush_Interval() {
	done := make(chan bool, 1)
	buff := buffer.NewBuffer[string]()
	buff.SetLimit(10000). // this amount ensure the flush trigger by interval
				SetInterval(10*time.Millisecond).
				SetConsumeFunc(true, func(items []string) {
			defer buff.Stop()
			defer func() { done <- true }()
			if s.Assert().Equal(len(items), 3) {
				for i := range items {
					s.Assert().True(strings.HasPrefix(items[i], "test_item_"))
				}
			}
		}).
		Start(context.Background())

	timeout := time.After(20 * time.Millisecond)
	wg := new(sync.WaitGroup)
	wg.Add(1)
	go func() {
		defer wg.Done()
		select {
		case <-done:
			return
		case <-timeout:
			s.FailNow("buffer should flush before timeout")
		}
	}()

	buff.Write("test_item_1")
	buff.Write("test_item_2")
	buff.Write("test_item_3")

	wg.Wait()
}
