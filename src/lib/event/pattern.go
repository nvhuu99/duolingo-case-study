package event

type Pattern interface {
	IsEmptyPattern() bool
	Match(target Pattern) bool
	Equal(p Pattern) bool
}
