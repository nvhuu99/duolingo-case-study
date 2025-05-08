package rest_http

import (
	"encoding/json"
	"net/http"
)

type Response struct {
	rw http.ResponseWriter

	Status           int    `json:"status"`
	Success          bool   `json:"success"`
	Message          string `json:"message"`
	Data             any    `json:"data"`
	Errors           any    `json:"errors"`
	ResponseTimeMs   int    `json:"-"`
	ResponseBodySize int    `json:"-"`
	ResponseSent     bool   `json:"-"`
}

func NewResponse(rw http.ResponseWriter) *Response {
	return &Response{rw: rw}
}

func (response *Response) Ok(message string, data any) *Response {
	response.Status = STATUS_OK
	response.Success = true
	response.Message = message
	if data != nil {
		response.Data = data
	}
	return response
}

func (response *Response) Created(message string, data any) *Response {
	response.Status = STATUS_CREATED
	response.Success = true
	response.Message = message
	if data != nil {
		response.Data = data
	}
	return response
}

func (response *Response) NotFound(message string) *Response {
	response.Status = STATUS_NOT_FOUND
	response.Success = true
	response.Message = message
	return response
}

func (response *Response) InvalidRequest(message string, errs any) *Response {
	response.Status = STATUS_INVALID
	response.Success = true
	response.Message = message
	if errs != nil {
		response.Errors = errs
	}
	return response
}

func (response *Response) ServerErr(message string, errs any) *Response {
	response.Status = STATUS_SERVER_ERR
	response.Success = true
	response.Message = message
	if errs != nil {
		response.Errors = errs
	}
	return response
}

func (response *Response) Header() http.Header {
	return response.rw.Header()
}

func (response *Response) GetBody() []byte {
	body, _ := json.Marshal(response)
	return body
}
