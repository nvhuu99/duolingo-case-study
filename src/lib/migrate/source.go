package migrate

// Source is responsible for reading the database migration files 
// and creating Migration objects. Then Migration objects will be 
// executed by the Database driver.
type Source interface {
	Open(batch []string) error
	List() []string
	HasNext() bool
	Next() (*Migration, error)
	Close()
	HasError() bool
	Error() error
}