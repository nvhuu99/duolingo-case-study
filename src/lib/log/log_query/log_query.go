package log_query

import (
	"duolingo/lib/log"
	lg "duolingo/lib/log"
)

type LogQuery interface {
	Info() LogQuery

	Error() LogQuery

	Debug() LogQuery

	Filters(conditions map[string]any) LogQuery

	Sum(extractor func(*lg.Log) float64) (float64, error)

	Avg(extractor func(*lg.Log) float64) (float64, error)

	First(filter func(*log.Log) bool) (*log.Log, error)

	Count(filter func(*lg.Log) bool) (int, error)

	Any(filter func(*lg.Log) bool) (bool, error)

	Each(callback func(*lg.Log) LoopAction) error

	All() ([]*lg.Log, error)
}
