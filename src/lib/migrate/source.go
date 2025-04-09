package migrate

// Source is responsible for reading the database migration files
// and creating Migration objects. These Migration objects are then
// executed by a database driver.
type Source interface {
	UseUri(uri string) error
	Open(batch []string) error
	List() []string
	HasNext() bool
	Next() (*Migration, error)
	Close()
	HasError() bool
	Error() error
}
