package test_suites

import (
	"duolingo/libraries/work_distributor"

	"github.com/stretchr/testify/suite"
)

type AssignmentTestSuite struct {
	suite.Suite
}

func NewAssignmentTestSuite() *AssignmentTestSuite {
	return &AssignmentTestSuite{}
}

func (s *AssignmentTestSuite) Test_NewAssignment() {
	a1, err1 := work_distributor.NewAssignment("", "", 0, 0)      // missing ids
	a2, err2 := work_distributor.NewAssignment("A2", "", 0, 0)    // missing workload id
	a3, err3 := work_distributor.NewAssignment("A3", "W1", 10, 1) // end < start
	s.Assert().Nil(a1)
	s.Assert().Nil(a2)
	s.Assert().Nil(a3)
	s.Assert().Error(err1)
	s.Assert().Error(err2)
	s.Assert().Error(err3)

	a4, err4 := work_distributor.NewAssignment("A4", "W1", 1, 10)
	s.Assert().NotNil(a4)
	s.Assert().NoError(err4)
}

func (s *AssignmentTestSuite) Test_WorkStartAt_WorkEndAt() {
	assignment, _ := work_distributor.NewAssignment("A1", "W1", 1, 100)

	s.Assert().Equal(uint64(1), assignment.WorkStartAt())
	s.Assert().Equal(uint64(100), assignment.WorkEndAt())

	assignment.Progress = 50

	s.Assert().Equal(uint64(51), assignment.WorkStartAt())
}
