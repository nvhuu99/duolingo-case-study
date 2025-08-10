package restful

import (
	"encoding/json"
	"errors"
	"net/http"
)

type Response struct {
	base http.ResponseWriter
	sent bool
	status int
	err error
	success bool
}

func NewResponse(base http.ResponseWriter) *Response {
	return &Response{base: base}
}

func (res *Response) Sent() bool { return res.sent }
func (res *Response) Status() int { return res.status }
func (res *Response) Success() bool { return res.success }

func (res *Response) Error() error { return res.err }
func (res *Response) SetErr(errs any) {
	if asErr, ok := errs.(error); ok {
		res.err = asErr
		return
	} else {
		if asString, err := json.Marshal(errs); err == nil {
			res.err = errors.New(string(asString))
		}
	}
}

func (res *Response) SetHeader(key string, value string) {
	res.base.Header().Set(key, value)
}

func (res *Response) Ok(message string, data any) {
	res.Send(http.StatusOK, true, message, nil, data)
}

func (res *Response) Created(message string, data any) {
	res.Send(http.StatusCreated, true, message, nil, data)
}

func (res *Response) NotFound(message string) {
	res.Send(http.StatusNotFound, false, message, errors.New(message), nil)
}

func (res *Response) BadRequest(message string, errs any) {
	res.Send(http.StatusBadRequest, false, message, errs, nil)
}

func (res *Response) ServerErr(message string) {
	res.Send(http.StatusInternalServerError, false, message, errors.New(message), nil)
}

func (res *Response) NoContent() {
	res.Send(http.StatusNoContent, true, "", nil, nil)
}

func (res *Response) Send(
	status int,
	success bool,
	message string,
	errors any,
	data any,
) {
	if res.sent {
		return
	}

	body, bodyErr := res.buildBody(status, success, message, errors, data)
	if bodyErr != nil {
		panic(bodyErr)
	}

	res.base.Header().Set("Content-Type", "application/json")
	res.base.WriteHeader(status)
	res.base.Write(body)

	res.sent = true
	res.status = status
	res.success = success
	res.SetErr(errors)
}

func (res *Response) buildBody(
	status int,
	success bool,
	message string,
	errors any,
	data any,
) ([]byte, error) {
	body := map[string]any{
		"status":  status,
		"success": success,
		"message": message,
		"errors":  errors,
		"data":    data,
	}
	return json.Marshal(body)
}


