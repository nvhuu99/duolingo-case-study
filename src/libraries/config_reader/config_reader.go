package config_reader

var (
	ErrSourceIsNotSet = "no config source registered for %v"
	ErrSourceFailure  = "%v source error: %v"
	ErrConfigNotFound = "%v not found from source %v"
)

type Source interface {
	Load(uri string) ([][]byte, error)
}

type ConfigReader interface {
	Get(uri string, pattern string) string
	GetInt(uri string, pattern string) int
	GetInt64(uri string, pattern string) int64
	GetArr(uri string, pattern string) []string
	GetIntArr(uri string, pattern string) []int
	GetInt64Arr(uri string, pattern string) []int64
}
