package test_suites

import (
	"duolingo/libraries/work_distributor"

	"github.com/stretchr/testify/suite"
)

type WorkDistributorTestSuite struct {
	suite.Suite
	distributor work_distributor.WorkDistributor
}

func (s *WorkDistributorTestSuite) TestCreateEmptyWorkload() {
	workload, err := s.distributor.CreateWorkload(0)
	s.Assert().Nil(workload)
	s.Assert().Error(err)
}

func (s *WorkDistributorTestSuite) TestCreateWorkload() {
	workload, err := s.distributor.CreateWorkload(100)
	s.Assert().NotNil(workload)
	s.Assert().NoError(err)
}
