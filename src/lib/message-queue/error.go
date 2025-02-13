package messagequeue

import (
	"fmt"
	"runtime"
)

const (
	ConnectionFailure			= 501
	ConnectionTimeOut			= 502
	PublishFailure				= 503
	PublishConfirmFailure		= 504
	PublishNACK					= 505
	PublishTimeOutExceed		= 506
	DeclareFailure				= 507
	DeclareTimeOutExceed		= 508
	TopicDeclareFailure			= 509
	QueueDeclareFailure			= 510
	BindingDeclareFailure		= 511
	TopologyFailure				= 512
	ClientFatalError			= 513
	ManagerConfigMissing		= 514
)

var ErrMessages = map[int]string{
	ConnectionFailure:		"connection failure",
	ConnectionTimeOut:		"connection timeout",
	PublishFailure:			"publish message failure",
	PublishConfirmFailure:	"publish message confirm failure",
	PublishNACK:			"publish message not acknowledged (NACK)",
	PublishTimeOutExceed:	"publish message timeout exceeded",
	DeclareTimeOutExceed:	"declare timeout exceeded",
	TopicDeclareFailure:	"topic declare failure",
	QueueDeclareFailure:	"queue declare failure",
	BindingDeclareFailure:	"binding beclare failure",
	TopologyFailure:		"topology operation failure",
	ClientFatalError:		"client operations fatal error",
	ManagerConfigMissing:	"manager configuration missing",
}

type Error struct {
	Topic  string
	Queue string
	Pattern string
	
	Code          int
	Message       string
	OriginalError error
	
	FuncName   string
	File       string
	LineNumber int
}

func (e *Error) Error() string {
	mssg := fmt.Sprintf(
		"file: \"%v\", line: %v, function: \"%v\"\ncode: %v, error message: \"%v\"\n",
		e.File,
		e.LineNumber,
		e.FuncName,
		e.Code,
		e.Message,
	)

	if e.Topic != "" {
		mssg += "topic: " + e.Topic
		if e.Queue != "" {
			mssg += ", queue: " + e.Queue
		}
		if e.Pattern != "" {
			mssg += ", pattern: " + e.Pattern
		}
		mssg += "\n"
	}

	if e.OriginalError != nil {
		mssg += fmt.Sprintf("original error: %v", e.OriginalError)
	}

	return mssg
}

func NewError(code int, err error, topic string, queue string, pattern string) *Error {
	mqErr := Error {
		Code: code,
		Message: ErrMessages[code],
		OriginalError: err,
		Topic: topic,
		Queue: queue,
		Pattern: pattern,
	}
	pc, file, line, ok := runtime.Caller(1)
	if ok {
		mqErr.File = file
		mqErr.LineNumber = line
	}
	if f := runtime.FuncForPC(pc); f != nil {
		mqErr.FuncName = f.Name()
	}

	return &mqErr
}