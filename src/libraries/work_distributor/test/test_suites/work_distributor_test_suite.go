package test_suites

import (
	"context"
	distributor "duolingo/libraries/work_distributor"
	"errors"
	"sync"
	"time"

	"github.com/stretchr/testify/suite"
)

type WorkDistributorTestSuite struct {
	suite.Suite
	distributor *distributor.WorkDistributor
}

func NewWorkDistributorTestSuite(distributor *distributor.WorkDistributor) *WorkDistributorTestSuite {
	return &WorkDistributorTestSuite{
		distributor: distributor,
	}
}

func (s *WorkDistributorTestSuite) Test_CreateWorkload_GetWorkload_And_AssignAll() {
	workload, err := s.distributor.CreateWorkload(100)
	getResult, getErr := s.distributor.GetWorkload(workload.Id)
	defer s.distributor.DeleteWorkloadAndAssignments(workload.Id)

	s.Assert().NotNil(workload)
	s.Assert().NoError(err)
	s.Assert().True(workload.Equal(getResult))
	s.Assert().NoError(getErr)

	total := workload.GetExpectTotalAssignments()
	assignments := []*distributor.Assignment{}
	for {
		assigned, assignErr := s.distributor.Assign(workload.Id)
		if assigned == nil || assignErr != nil {
			break
		}
		s.distributor.Commit(assigned)
		assignments = append(assignments, assigned)
	}

	if !s.Assert().Equal(total, int64(len(assignments))) {
		return
	}

	size := s.distributor.GetDistributionSize()
	for i := range total {
		start := i*size + 1
		end := start + size - 1
		s.Assert().Equal(workload.Id, assignments[i].WorkloadId)
		s.Assert().Equal(int64(start), assignments[i].StartIndex)
		s.Assert().Equal(int64(end), assignments[i].EndIndex)
		s.Assert().Zero(assignments[i].Progress)
	}
}

func (s *WorkDistributorTestSuite) Test_CommitAssignment_And_HasWorkloadFulfilled() {
	workload, _ := s.distributor.CreateWorkload(100)
	defer s.distributor.DeleteWorkloadAndAssignments(workload.Id)

	total := workload.GetExpectTotalAssignments()
	assignments := []*distributor.Assignment{}
	for {
		// if HasWorkloadFulfilled not work, later assertion would fail
		isFulfilled, isFulfilledErr := s.distributor.HasWorkloadFulfilled(workload.Id)
		if isFulfilled {
			break
		}
		if !s.Assert().NoError(isFulfilledErr) {
			break
		}
		// store the assignment
		assigned, assignErr := s.distributor.Assign(workload.Id)
		if assigned == nil && assignErr == distributor.ErrWorkloadHasAlreadyFulfilled {
			break
		}
		assignments = append(assignments, assigned)
		// commit the assignment
		commitErr := s.distributor.Commit(assigned)
		s.Assert().NoError(commitErr)
	}

	if s.Assert().Equal(total, int64(len(assignments))) {
		isFulfilled, isFulfilledErr := s.distributor.HasWorkloadFulfilled(workload.Id)
		s.Assert().True(isFulfilled)
		s.Assert().NoError(isFulfilledErr)
	}
}

func (s *WorkDistributorTestSuite) Test_CommitProgress_And_Rollback() {
	workload, _ := s.distributor.CreateWorkload(100)
	defer s.distributor.DeleteWorkloadAndAssignments(workload.Id)

	targetAssigment, _ := s.distributor.Assign(workload.Id)
	currentProgress := targetAssigment.Progress
	newProgress := currentProgress + 1

	commitErr := s.distributor.CommitProgress(targetAssigment, newProgress)
	rollbackErr := s.distributor.Rollback(targetAssigment)

	s.Assert().NoError(commitErr)
	s.Assert().NoError(rollbackErr)

	var rollbackedFound *distributor.Assignment
	for {
		assigned, assignErr := s.distributor.Assign(workload.Id)
		if assigned == nil || assignErr != nil {
			break
		}
		if assigned.Equal(targetAssigment) {
			rollbackedFound = assigned
			break
		}
	}

	s.Assert().NotNil(rollbackedFound)
	s.Assert().Equal(newProgress, rollbackedFound.Progress)
}

func (s *WorkDistributorTestSuite) Test_WaitForAssignment_WaitUntilOneRollbacked() {
	workload, _ := s.distributor.CreateWorkload(100)
	defer s.distributor.DeleteWorkloadAndAssignments(workload.Id)

	// Distribute all assignments
	var assigned *distributor.Assignment
	for {
		newAssignment, _ := s.distributor.Assign(workload.Id)
		if newAssignment == nil {
			break
		}
		assigned = newAssignment
	}

	// Create wait group to wait for all goroutines
	var wg sync.WaitGroup
	wg.Add(2)

	// Create a timeout to wait for available assignment
	waitCtx, waitCancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer waitCancel()

	// Call WaitForAssignment() to wait for available work
	var assignedByWait *distributor.Assignment
	var assignErr error
	go func() {
		defer wg.Done()
		assignedByWait, assignErr = s.distributor.WaitForAssignment(
			waitCtx,
			5*time.Millisecond,
			workload.Id,
		)
	}()

	// Rollbacked an assignment,
	// then verify that eventually the work is available
	go func() {
		defer wg.Done()
		rollbackTimer := time.After(10 * time.Millisecond)
		for {
			select {
			case <-rollbackTimer:
				s.distributor.Rollback(assigned)
			case <-waitCtx.Done():
				s.Assert().NoError(assignErr)
				s.Assert().NotNil(assignedByWait)
				return
			}
		}
	}()

	wg.Wait()
}

func (s *WorkDistributorTestSuite) Test_WaitForAssignment_WaitWillFailOnFulfilled() {
	workload, _ := s.distributor.CreateWorkload(100)
	defer s.distributor.DeleteWorkloadAndAssignments(workload.Id)

	// Distribute all assignments
	assignments := []*distributor.Assignment{}
	for {
		newAssignment, _ := s.distributor.Assign(workload.Id)
		if newAssignment == nil {
			break
		}
		assignments = append(assignments, newAssignment)
	}

	// Create wait group to wait for all goroutines
	var wg sync.WaitGroup
	wg.Add(2)

	// Create a timeout to wait for available assignment
	waitCtx, waitCancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer waitCancel()

	// Call WaitForAssignment() to wait for available work
	var assignedByWait *distributor.Assignment
	var assignErr error
	go func() {
		defer wg.Done()
		assignedByWait, assignErr = s.distributor.WaitForAssignment(
			waitCtx,
			5*time.Millisecond,
			workload.Id,
		)
	}()

	// Commit all assignments,
	// then verify that there no more work
	go func() {
		defer wg.Done()
		for i := range assignments {
			s.distributor.Commit(assignments[i])
		}
		time.Sleep(10 * time.Millisecond)
		<-waitCtx.Done()
		s.Assert().Nil(assignedByWait)
		s.Assert().Equal(distributor.ErrWorkloadHasAlreadyFulfilled, assignErr)
	}()

	wg.Wait()
}

func (s *WorkDistributorTestSuite) Test_HandleAssignment() {
	workload, _ := s.distributor.CreateWorkload(20)
	defer s.distributor.DeleteWorkloadAndAssignments(workload.Id)

	// Steps:
	// 1. First assignment             - Remains: assignment2
	// 2. Handle failed & rollbacked   - Remains: assignment2, assignment1
	// 3. Second assignment            - Remains: assignment1
	// 4. Handle succeeded & committed - Remains: assignment1
	// 5. First assignment re-assigned - Remains: empty
	// 6. Handle succeeded, workload fulfilled

	assignment1, _ := s.distributor.Assign(workload.Id)
	s.distributor.HandleAssignment(assignment1, func() error {
		return errors.New("stimulate failure")
	})
	assignment2, _ := s.distributor.Assign(workload.Id)
	s.distributor.HandleAssignment(assignment2, func() error {
		return nil
	})
	assignment3, _ := s.distributor.Assign(workload.Id)
	s.distributor.HandleAssignment(assignment3, func() error {
		return nil
	})

	// ensure first assignment rollbacked
	s.Assert().True(assignment3.Equal(assignment1))
	// ensure all assignments committed
	s.Assert().True(s.distributor.HasWorkloadFulfilled(workload.Id))
}
