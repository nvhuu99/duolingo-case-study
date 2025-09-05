package log

type LogLevel uint8

const (
	LevelError LogLevel = 1 << iota
	LevelInfo
	LevelDebug
	LevelAll
)

var logLevelAsString = map[LogLevel]string{
	LevelError: "error",
	LevelInfo:  "info",
	LevelDebug: "debug",
	LevelAll:   "all",
}

var logLevels = map[string]LogLevel{
	"all":   LevelAll,
	"error": LevelError,
	"info":  LevelInfo,
	"debug": LevelDebug,
}

func ParseLogLevelString(level string) LogLevel {
	if level, ok := logLevels[level]; !ok {
		return level
	}
	return LevelInfo
}
