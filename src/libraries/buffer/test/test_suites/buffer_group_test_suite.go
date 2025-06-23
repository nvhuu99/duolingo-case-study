package test_suites

import (
	"context"
	"duolingo/libraries/buffer"
	"strings"
	"sync"
	"time"

	"github.com/stretchr/testify/suite"
)

type BufferGroupTestSuite struct {
	suite.Suite
}

func (s *BufferGroupTestSuite) Test_BufferGroup_Limit() {
	done := make(chan bool, 1)

	flushCount := 0
	grp := buffer.NewBufferGroup[string, string](context.Background())
	grp.
		SetLimit(3).
		SetInterval(100*time.Second). // this amount ensure the flush trigger by limit
		AddGroup("grp_1").
		AddGroup("grp_2").
		SetConsumeFunc(true, func(name string, items []string) {
			if s.Assert().True(name == "grp_1" || name == "grp_2") {
				if s.Assert().Equal(len(items), 3) {
					for i := range items {
						s.Assert().True(strings.HasPrefix(items[i], "test_item_"))
					}
					if flushCount++; flushCount == 2 {
						done <- true
					}
				}
			}
		})

	grp.Start()
	defer grp.Stop()

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

	grp.Write("grp_1", "test_item_1")
	grp.Write("grp_1", "test_item_2")
	grp.Write("grp_1", "test_item_3")
	grp.Write("grp_1", "test_item_4") // 3 items limit hit, should trigger flushing

	grp.Write("grp_2", "test_item_1")
	grp.Write("grp_2", "test_item_2")
	grp.Write("grp_2", "test_item_3")
	grp.Write("grp_2", "test_item_4") // 3 items limit hit, should trigger flushing

	wg.Wait()
}

func (s *BufferGroupTestSuite) Test_BufferGroup_Flush_Interval() {
	done := make(chan bool, 1)

	flushCount := 0
	grp := buffer.NewBufferGroup[string, string](context.Background())
	grp.
		SetInterval(10*time.Millisecond).
		SetLimit(1000). // this amount ensure the flush trigger by interval
		AddGroup("grp_1").
		AddGroup("grp_2").
		SetConsumeFunc(true, func(name string, items []string) {
			if s.Assert().True(name == "grp_1" || name == "grp_2") {
				if s.Assert().Equal(3, len(items)) {
					for i := range items {
						s.Assert().True(strings.HasPrefix(items[i], "test_item_"))
					}
					if flushCount++; flushCount == 2 {
						done <- true
					}
				}
			}
		})

	grp.Start()
	defer grp.Stop()

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

	grp.Write("grp_1", "test_item_1")
	grp.Write("grp_1", "test_item_2")
	grp.Write("grp_1", "test_item_3")

	grp.Write("grp_2", "test_item_1")
	grp.Write("grp_2", "test_item_2")
	grp.Write("grp_2", "test_item_3")

	wg.Wait()
}
