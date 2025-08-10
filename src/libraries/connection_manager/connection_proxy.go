package connection_manager

import (
	"errors"
	"io"
	"net"
	"strings"
	"syscall"
)

type ConnectionProxy interface {
	SetArgsPanicIfInvalid(args any)
	MakeConnection() (any, error)
	Ping(connection any) error
	IsNetworkErr(err error) bool
	CloseConnection(connection any)
	ConnectionName() string
}

// A fallback network error check for ConnectionProxy implementations
func IsNetworkErr(err error) bool {
	if err == nil {
		return false
	}
	// io.EOF or io.ErrUnexpectedEOF
	if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
		return true
	}
	// net.Error (timeout, temporary, etc.)
	var netErr net.Error
	if errors.As(err, &netErr) {
		return true
	}
	// syscall.Errno (e.g. ECONNRESET)
	var syscallErr syscall.Errno
	if errors.As(err, &syscallErr) {
		return true
	}
	// Some network-level errors are only visible via message content
	msg := err.Error()
	if strings.Contains(msg, "broken pipe") ||
		strings.Contains(msg, "connection reset") ||
		strings.Contains(msg, "i/o timeout") ||
		strings.Contains(msg, "connection refused") ||
		strings.Contains(msg, "connection closed") {
		return true
	}
	return false
}
