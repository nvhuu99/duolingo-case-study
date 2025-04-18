package event_data

import (
	wd "duolingo/lib/work_distributor"
	"duolingo/model"
)

type BuildPushNotiMessage struct {
	OptId       string
	PushNoti    *model.PushNotiMessage
	Workload    *wd.Workload
	Assignments []*wd.Assignment
	Success     bool
	Error       error
}
