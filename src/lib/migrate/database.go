package migrate

type Database interface {
	PrepareDatabase() error
	BatchNumber() int
	LastBatch() ([]Migration, error)
	RunMigration(migr *Migration) error
	SaveMigrationRecord(migr *Migration) error
	DeleteMigrationRecord(migr *Migration) error
}
