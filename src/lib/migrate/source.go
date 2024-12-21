package migrate

// Source is responsible for reading the database migration files 
// and creating Migration objects. Then Migration objects will be 
// executed by the Database driver.
type Source interface {
	Open(ver string) error
	List() map[string]string
	HasNext() bool
	HasPrev() bool
	Next() (*Migration, error)
	Prev() (*Migration, error)
	Close()
	HasError() bool
	Error() error
}