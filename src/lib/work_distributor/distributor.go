package work_distributor

type Distributor interface {
	SetConnection(host string, port string) error
	PurgeData() error

	WorkloadExists(workloadName string) (bool, error)
	RegisterWorkLoad(workload *Workload) error
	SwitchToWorkload(workload string) (*Workload, error)

	Next() (*Assignment, error)
	Progress(assignmentId string, newVal int) error
	Commit(assignmentId string) error
	RollBack(assignmentId string) error
}
