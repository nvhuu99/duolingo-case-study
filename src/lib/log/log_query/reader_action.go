package log_query

type LoopAction string

const (
	LoopContinue LoopAction = "loop_continue"
	LoopCancel   LoopAction = "loop_cancel"
)
