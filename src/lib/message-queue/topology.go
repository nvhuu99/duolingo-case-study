package messagequeue

type Topology interface {
	Declare() *Error
	CleanUp() *Error
	IsReady() bool
}