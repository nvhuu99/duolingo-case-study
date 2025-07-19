package config_reader

var (
	ErrSourceNotRegistered = "trying to read an unregistered source %v"
	ErrSourceFailure       = "%v source error: %v"
	ErrConfigNotFound      = "%v not found from source %v"
)

type Source interface {
	Load() ([]byte, error)
}

type ConfigReader interface {
	Get(source string, pattern string) string
	GetInt(source string, pattern string) int
	GetInt64(source string, pattern string) int64
	GetArr(source string, pattern string) []string
	GetIntArr(source string, pattern string) []int
	GetInt64Arr(source string, pattern string) []int64
}
