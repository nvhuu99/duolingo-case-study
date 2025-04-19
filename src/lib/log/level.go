package log

type LogLevel uint8

const (
	LevelInfo LogLevel = 1 << iota
	LevelDebug
	LevelError
	LevelAll = LevelInfo | LevelDebug | LevelError
)

var LogLevelAsString = map[LogLevel]string{
	LevelInfo:  "info",
	LevelDebug: "debug",
	LevelError: "error",
}
