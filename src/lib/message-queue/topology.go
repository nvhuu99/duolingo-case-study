package messagequeue

type Topology interface {
	Declare() error
	CleanUp() error
	IsReady() bool
}