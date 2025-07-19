package restful

import (
	"encoding/json"
	"net/http"
)

type Response struct {
	base http.ResponseWriter
	sent bool
}

func NewResponse(base http.ResponseWriter) *Response {
	return &Response{base: base}
}

func (res *Response) Sent() bool {
	return res.sent
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
	res.Send(http.StatusNotFound, false, message, nil, nil)
}

func (res *Response) BadRequest(message string, errs any) {
	res.Send(http.StatusBadRequest, false, message, errs, nil)
}

func (res *Response) ServerErr(message string) {
	res.Send(http.StatusInternalServerError, false, "", nil, nil)
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

	body, err := res.buildBody(status, success, message, errors, data)
	if err != nil {
		panic(err)
	}

	res.base.Header().Set("Content-Type", "application/json")
	res.base.WriteHeader(status)
	res.base.Write(body)

	res.sent = true
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
