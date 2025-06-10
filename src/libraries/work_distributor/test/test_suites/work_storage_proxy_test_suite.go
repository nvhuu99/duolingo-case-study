package test_suites

import (
	"duolingo/libraries/work_distributor"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
)

type WorkStorageProxyTestSuite struct {
	suite.Suite
	proxy work_distributor.WorkStorageProxy
}

func NewWorkStorageProxyTestSuite(proxy work_distributor.WorkStorageProxy) *WorkStorageProxyTestSuite {
	return &WorkStorageProxyTestSuite{
		proxy: proxy,
	}
}

func (s *WorkStorageProxyTestSuite) Test_SaveWorkload() {
	workload, _ := work_distributor.NewWorkload(uuid.NewString(), 100, 10)
	saveErr := s.proxy.SaveWorkload(workload)
	getResult, _ := s.proxy.GetWorkload(workload.Id)
	defer s.proxy.DeleteWorkloadAndAssignments(workload.Id)

	s.Assert().NoError(saveErr)
	s.Assert().True(workload.Equal(getResult))
}

func (s *WorkStorageProxyTestSuite) Test_GetWorkload() {
	workload, _ := work_distributor.NewWorkload(uuid.NewString(), 100, 10)
	s.proxy.SaveWorkload(workload)
	defer s.proxy.DeleteWorkloadAndAssignments(workload.Id)

	getResult1, getErr1 := s.proxy.GetWorkload("not_exist_id")
	getResult2, getErr2 := s.proxy.GetWorkload(workload.Id)

	s.Assert().Nil(getResult1)
	s.Assert().Equal(work_distributor.ErrWorkloadNotExists, getErr1)

	s.Assert().True(workload.Equal(getResult2))
	s.Assert().NoError(getErr2)
}

func (s *WorkStorageProxyTestSuite) Test_GetAndUpdateWorkload() {

	workload, _ := work_distributor.NewWorkload(uuid.NewString(), 100, 10)
	s.proxy.SaveWorkload(workload)
	defer s.proxy.DeleteWorkloadAndAssignments(workload.Id)

	saveErr := s.proxy.GetAndUpdateWorkload(workload.Id, func(w *work_distributor.Workload) error {
		w.TotalCommittedAssignments = w.GetExpectTotalAssignments()
		return nil
	})
	getResult, _ := s.proxy.GetWorkload(workload.Id)
	s.Assert().NoError(saveErr)
	s.Assert().True(getResult.HasWorkloadFulfilled())
}

func (s *WorkStorageProxyTestSuite) Test_PushAndPop_AssignmentToQueue() {
	workload, _ := work_distributor.NewWorkload(uuid.NewString(), 100, 10)
	s.proxy.SaveWorkload(workload)
	defer s.proxy.DeleteWorkloadAndAssignments(workload.Id)

	assignment1, _ := work_distributor.NewAssignment("a1", workload.Id, 1, 10)
	assignment2, _ := work_distributor.NewAssignment("a2", workload.Id, 1, 10)

	pushErr1 := s.proxy.PushAssignmentToQueue(assignment1)
	pushErr2 := s.proxy.PushAssignmentToQueue(assignment2)
	if !s.Assert().NoError(pushErr1) || !s.Assert().NoError(pushErr2) {
		return
	}

	poppedResult1, popErr1 := s.proxy.PopAssignmentFromQueue(workload.Id)
	poppedResult2, popErr2 := s.proxy.PopAssignmentFromQueue(workload.Id)
	if !s.Assert().NoError(popErr1) || !s.Assert().NoError(popErr2) {
		return
	}
	s.Assert().True(assignment1.Equal(poppedResult1))
	s.Assert().True(assignment2.Equal(poppedResult2))

	poppedResult3, popErr3 := s.proxy.PopAssignmentFromQueue(workload.Id)
	s.Assert().Nil(poppedResult3)
	s.Assert().NoError(popErr3)
}

func (s *WorkStorageProxyTestSuite) Test_DeleteWorkloadAndAssignments() {
	workload, _ := work_distributor.NewWorkload(uuid.NewString(), 100, 10)
	s.proxy.SaveWorkload(workload)

	assignment1, _ := work_distributor.NewAssignment("a1", workload.Id, 1, 10)
	assignment2, _ := work_distributor.NewAssignment("a2", workload.Id, 1, 10)
	s.proxy.PushAssignmentToQueue(assignment1)
	s.proxy.PushAssignmentToQueue(assignment2)

	delErr := s.proxy.DeleteWorkloadAndAssignments(workload.Id)
	_, getErr := s.proxy.GetWorkload(workload.Id)
	_, popErr := s.proxy.PopAssignmentFromQueue(workload.Id)

	s.Assert().NoError(delErr)
	s.Assert().Equal(work_distributor.ErrWorkloadNotExists, getErr)
	s.Assert().Error(work_distributor.ErrWorkloadNotExists, popErr)
}
