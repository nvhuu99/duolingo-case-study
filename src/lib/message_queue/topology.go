package message_queue

type Topology interface {
	Declare() error
	CleanUp() error
	IsReady() bool
}
