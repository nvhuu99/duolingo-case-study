package migrate

type Database interface {
	SetConnection(host string, port string, usr string, pwd string)
	SetDatabase(database string)
	GetFileExt() string
	PrepareDatabase() error
	BatchNumber() int
	LastBatch() ([]Migration, error)
	RunMigration(migr *Migration) error
	SaveMigrationRecord(migr *Migration) error
	DeleteMigrationRecord(migr *Migration) error
}
