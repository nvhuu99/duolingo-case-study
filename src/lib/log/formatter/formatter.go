package formatter

type Formatter interface {
	Format(any) []byte
}
