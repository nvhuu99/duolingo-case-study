package migrate

type MigrateType string

type MigrateStatus string

const (
	NilVersion      string      = "-1"
	MigrateUp       MigrateType = "up"
	MigrateRollback MigrateType = "rollback"

	MigrateRunning  MigrateStatus = "running"
	MigrateFinished MigrateStatus = "finished"
	MigrateFailed   MigrateStatus = "failed"
)
