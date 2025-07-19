package test_suites

import (
	distributor "duolingo/libraries/work_distributor"

	"github.com/stretchr/testify/suite"
)

type WorkloadTestSuite struct {
	suite.Suite
}

func NewWorkloadTestSuite() *WorkloadTestSuite {
	return &WorkloadTestSuite{}
}

func (s *WorkloadTestSuite) Test_NewWorkload() {
	w1, err1 := distributor.NewWorkload("", 0, 0)     // missing all params
	w2, err2 := distributor.NewWorkload("W2", 100, 0) // missing distribution size

	s.Assert().Nil(w1)
	s.Assert().Nil(w2)
	s.Assert().Error(err1)
	s.Assert().Error(err2)

	// distribution size's greater than total units
	w3, err3 := distributor.NewWorkload("W3", 100, 1000)
	// normal case
	w4, err4 := distributor.NewWorkload("W4", 100, 5)

	s.Assert().NotNil(w3)
	s.Assert().NotNil(w4)
	s.Assert().NoError(err3)
	s.Assert().NoError(err4)
}

func (s *WorkloadTestSuite) Test_GetExpectTotalAssignments() {
	w1, _ := distributor.NewWorkload("W1", 100, 10)
	w2, _ := distributor.NewWorkload("W2", 100, 3)

	s.Assert().Equal(int64(10), w1.GetExpectTotalAssignments())
	s.Assert().Equal(int64(34), w2.GetExpectTotalAssignments())
}

func (s *WorkloadTestSuite) Test_HasWorkloadFulfilled() {
	w1, _ := distributor.NewWorkload("W1", 100, 10)

	w1.TotalCommittedAssignments = 5
	s.Assert().False(w1.HasWorkloadFulfilled())

	w1.TotalCommittedAssignments = 10
	s.Assert().True(w1.HasWorkloadFulfilled())
}

func (s *WorkloadTestSuite) Test_IncreaseTotalCommittedAssignments() {
	w1, _ := distributor.NewWorkload("W1", 100, 10)
	w1.TotalCommittedAssignments = 10
	err := w1.IncreaseTotalCommittedAssignments()

	s.Assert().Equal(distributor.ErrUnexpectedWorkloadTotalCommittedAssignments, err)
}
