package rest_http

import (
	"encoding/json"
	"runtime"
)

const (
	ERR_URL_DECODE_FAILURE = 501
	ERR_ARGUMENT_NOT_ENCLOSED = 502
	ERR_ROUTE_NOT_FOUND = 503
	ERR_SERVER_PANIC = 504
)

var ErrMessages = map[int]string{
	ERR_URL_DECODE_FAILURE: "request url contains invalid url-encode character",
	ERR_ARGUMENT_NOT_ENCLOSED: "route argument is not enclosed with \"{}\"",
	ERR_ROUTE_NOT_FOUND: "the requested route does not exist",
	ERR_SERVER_PANIC: "server panic",
}

type Error struct {
	Code	int		`json:"code"`
	Message	string	`json:"message"`
	
	Method	string	`json:"method"`
	Uri		string	`json:"uri"`
	
	FuncName   		string	`json:"funcName"`
	File       		string	`json:"file"`
	LineNumber 		int		`json:"lineNumber"`

	OriginalErr		any		`json:"-"`
	OriginalErrMsg	string	`json:"originalError"`
}

func (e *Error) Error() string {
	if originalErr, ok := e.OriginalErr.(error); ok {
		e.OriginalErrMsg = originalErr.Error()
	}
	
	err, _ := json.Marshal(e)

	return string(err)
}

func NewError(code int, err any, method string, uri string) *Error {
	mqErr := Error {
		Code: code,
		Message: ErrMessages[code],
		OriginalErr: err,
		Method: method,
		Uri: uri,
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