package migrate

type Database interface {
	SetConnectionString(uri string) Database
	SetDatabase(database string) Database
	SetConnection(host string, port string, usr string, pwd string) Database
	PrepareDatabase() error
	BatchNumber() int
	LastBatch() ([]Migration, error)
	RunMigration(migr *Migration) error
	SaveMigrationRecord(migr *Migration) error
	DeleteMigrationRecord(migr *Migration) error
}
