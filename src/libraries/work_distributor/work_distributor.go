package work_distributor

type WorkDistributor interface {
	CreateWorkload(totalWorkUnits uint64) (*Workload, error)
	// GetWorkload(workloadId string) (*Workload, error)
	// IsWorkloadFulfilled() bool
	// NextAssignment(workload *Workload, wait bool) *Assignment
	// Rollback(assignment *Assignment) error
	// CommitProgress(assignment *Assignment, workUnits int64) error
}
