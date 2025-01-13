package batchmanager

type BatchManager interface {
	SetConnection(host string, port string) error
	Reset() error

	NewBatch(id string) error
	Progress(id string, itemId string, val int) error
	Next(id string) (*BatchItem, error)
	Commit(id string, itemId string) error
	RollBack(id string, itemId string) error
}
