package notification

const (
	ERR_INVALID_CREDENTIALS = 501
	ERR_SEND_FAILURE        = 502
	ERR_DEVICE_TOKENS_EMPTY = 503
)

var ErrMessages = map[int]string{
	ERR_INVALID_CREDENTIALS: "501 - invalid credentials",
	ERR_SEND_FAILURE:        "502 - send failure",
	ERR_DEVICE_TOKENS_EMPTY: "503 - device tokens not provided",
}
