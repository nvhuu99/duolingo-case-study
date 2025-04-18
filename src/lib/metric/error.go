package metric

const (
	ERR_CAPTURE_STARTED_ALREADY = 501
	ERR_CAPTURE_HAS_NOT_STARTED = 502
	ERR_CAPTURE_ENDED           = 503
)

var ErrMessages = map[int]string{
	ERR_CAPTURE_STARTED_ALREADY: "501 - metrics capturing has already started",
	ERR_CAPTURE_HAS_NOT_STARTED: "502 - metrics capturing has not yet started",
	ERR_CAPTURE_ENDED:           "503 - metrics capturing ended",
}
