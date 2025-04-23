package writer

type LogOutput interface {
	Flush([]*Writable) error
}
