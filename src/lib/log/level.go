package log

type LogLevel string

const (
	LevelInfo    LogLevel = "info"
	LevelWarning LogLevel = "warning"
	LevelError   LogLevel = "error"
	LevelFatal   LogLevel = "fatal"
)

var levelFileExtensions = map[LogLevel]string{
	LevelInfo:    "info.json",
	LevelWarning: "warning.json",
	LevelError:   "error.json",
	LevelFatal:   "fatal.json",
}
