package test_suites

import (
	"duolingo/libraries/work_distributor"

	"github.com/stretchr/testify/suite"
)

type WorkloadTestSuite struct {
	suite.Suite
}

func (s *WorkloadTestSuite) TestCreateWorkloadWithInvalidParams() {
	w1, err1 := work_distributor.NewWorkload("", 0, 0)
	w2, err2 := work_distributor.NewWorkload("WORKLOAD#2", 100, 0)
	s.Assert().Nil(w1)
	s.Assert().Nil(w2)
	s.Assert().Error(err1)
	s.Assert().Error(err2)
}

func (s *WorkloadTestSuite) TestCreateWorkload() {
	w, err := work_distributor.NewWorkload("WORKLOAD#1", 100, 5)
	s.Assert().NotNil(w)
	s.Assert().NoError(err)
	s.Assert().Equal("WORKLOAD#1", w.GetId())
	s.Assert().Equal(100, w.GetTotalWorkloadUnits())
	s.Assert().Equal(5, w.GetDistributionSize())
}
