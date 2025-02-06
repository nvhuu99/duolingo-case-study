package messagequeue

type Topology interface {
	Declare() *Error
}