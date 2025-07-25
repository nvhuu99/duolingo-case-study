package work_distributor

import "errors"

var (
	ErrInvalidAssignment = errors.New("invalid assignment parameters")
)

type Assignment struct {
	Id         string `json:"id"`
	WorkloadId string `json:"workload_id"`
	StartIndex int64  `json:"start_idx"`
	EndIndex   int64  `json:"end_idx"`
	Progress   int64  `json:"progress"`
}

func NewAssignment(
	id string,
	workloadId string,
	startIdx int64,
	endIdx int64,
) (*Assignment, error) {
	assignment := &Assignment{
		Id:         id,
		WorkloadId: workloadId,
		StartIndex: startIdx,
		EndIndex:   endIdx,
		Progress:   0,
	}
	if err := assignment.Validate(); err != nil {
		return nil, err
	}
	return assignment, nil
}

func (assignment *Assignment) Validate() error {
	if assignment.Id == "" ||
		assignment.WorkloadId == "" ||
		assignment.StartIndex > assignment.EndIndex ||
		assignment.Progress > assignment.EndIndex {
		return ErrInvalidAssignment
	}
	return nil
}

func (assignment *Assignment) WorkStartAt() int64 {
	if assignment.Progress > assignment.StartIndex {
		return assignment.Progress + 1
	}
	return assignment.StartIndex
}

func (assignment *Assignment) WorkEndAt() int64 {
	return assignment.EndIndex
}

func (assignment *Assignment) IsCompleted() bool {
	return assignment.Progress == assignment.EndIndex
}

func (assignment *Assignment) Equal(target *Assignment) bool {
	return target != nil && assignment.Id == target.Id
}
