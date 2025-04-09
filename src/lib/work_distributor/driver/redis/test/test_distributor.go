package test

import (
	"context"
	wd "duolingo/lib/work_distributor"
	redis "duolingo/lib/work_distributor/driver/redis"
	"time"

	"github.com/stretchr/testify/suite"
)

type RedisDistributorTestSuite struct {
	suite.Suite

	Host string
	Port string

	distributor *redis.RedisDistributor
}

func (s *RedisDistributorTestSuite) SetupTest() {
	s.distributor, _ = redis.NewRedisDistributor(context.Background(), "redis-distributor-test")
	s.distributor.
		WithOptions(nil).
		WithLockTimeOut(3 * time.Second).
		WithDistributionSize(10)

	err := s.distributor.SetConnection(s.Host, s.Port)
	if err != nil {
		s.FailNow("connection failure")
	}
	s.distributor.PurgeData()
}

func (s *RedisDistributorTestSuite) TearDownTest() {
	s.distributor.PurgeData()
}

func (s *RedisDistributorTestSuite) TestRegisterWorkload() {
	name := "workload-test"
	exist, err := s.distributor.WorkloadExists(name)
	s.Require().False(exist, "workload must not exist before registered")
	s.Require().NoError(err, "WorkloadExists() should not error")

	err = s.distributor.RegisterWorkLoad(&wd.Workload{
		Name:             name,
		NumOfUnits:       100,
		DistributionSize: 10,
	})
	s.Require().NoError(err, "RegisterWorkLoad() should not error")

	exist, err = s.distributor.WorkloadExists(name)
	s.Require().True(exist, "workload must exist after registered")
	s.Require().NoError(err, "WorkloadExists() should not error")
}

func (s *RedisDistributorTestSuite) TestFullDistribution() {
	workload := &wd.Workload{
		Name:             "workload-test",
		NumOfUnits:       100,
		DistributionSize: 10,
	}
	s.distributor.RegisterWorkLoad(workload)
	s.distributor.SwitchToWorkload("workload-test")

	start := 1
	end := workload.DistributionSize
	for i := 0; i < workload.NumOfAssignments(); i++ {
		assignment, err := s.distributor.Next()
		s.Require().NoError(err, "Next() should not return an error")
		s.Require().Equal(assignment.Start, start, "assignment's start must be correct")
		s.Require().Equal(assignment.End, end, "assignment's start must be correct")

		s.distributor.Commit(assignment.Id)
		s.Require().NoError(err, "Commit() should not return an error")

		start = end + 1
		end = start + workload.DistributionSize - 1
	}

	assignment, _ := s.distributor.Next()
	s.Require().Nil(assignment, "nil assignment should be return after a complete distribution")
}

func (s *RedisDistributorTestSuite) TestRollbackAndResume() {
	workload := &wd.Workload{
		Name:             "workload-test",
		NumOfUnits:       100,
		DistributionSize: 10,
	}
	s.distributor.RegisterWorkLoad(workload)
	s.distributor.SwitchToWorkload("workload-test")
	// Skip first two assignments
	s.distributor.Next()
	s.distributor.Next()
	// Rollback the third assignment with half progress
	assignment, _ := s.distributor.Next()
	lastId := assignment.Id
	lastProgress := assignment.Start + 5
	err := s.distributor.Progress(assignment.Id, lastProgress)
	s.Require().NoError(err, "Progress() should not return an error")
	err = s.distributor.RollBack(assignment.Id)
	s.Require().NoError(err, "RollBack() should not return an error")
	// Validate re-assignment
	assignment, _ = s.distributor.Next()
	s.Require().Equal(assignment.Id, lastId, "the same assignment must be return after rollback")
	s.Require().Equal(assignment.Start, lastProgress+1, "the assignment's start must be one unit ahead of the last progress")
}
