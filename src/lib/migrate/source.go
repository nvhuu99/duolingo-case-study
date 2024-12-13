package migrate

type Source interface {
	Open(uri string) error
	List() (map[string]string, error)
	HasNext() bool
	HasPrev() bool
	Next() (*Migration, error)
	Prev() (*Migration, error)
	Close()
}