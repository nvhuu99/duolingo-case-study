package log

type Formatter interface {
	Format(any) []byte
}