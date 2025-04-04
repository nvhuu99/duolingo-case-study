package writer

type LogWriter interface {
	Write(log *Writable)
}