package migrate

type Database interface {
	GetVersion() string
	RunMigration(migr *Migration) error
}
