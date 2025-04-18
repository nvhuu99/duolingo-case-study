package metric

type CaptureFlag uint8
type CaptureStatus string

const (
	CaptureNone CaptureFlag = 0
	CaptureCPU  CaptureFlag = 1 << iota
	CaptureMemory
	CaptureDisksIO
	CaptureAll = CaptureCPU | CaptureMemory | CaptureDisksIO

	CaptureStatusStarted CaptureStatus = "started"
	CaptureStatusEnded   CaptureStatus = "ended"
)
