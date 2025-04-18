package log

type LogLevel uint8

const (
	LevelInfo  LogLevel = 0
	LevelDebug LogLevel = 1 << iota
	LevelError
	LevelAll = LevelInfo | LevelDebug | LevelError
)

var levelFileExtensions = map[LogLevel]string{
	LevelInfo:  "info",
	LevelDebug: "debug",
	LevelError: "error",
}
