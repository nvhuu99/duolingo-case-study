package resthttp

import (
	"encoding/json"
	"net/http"
)

type Response struct {
	writer http.ResponseWriter
}

func (res *Response) Ok(data any) {
	res.response(200, true, "OK", data, nil)
}

func (res *Response) Created(data any) {
	res.response(201, true, "Created", data, nil)
}

func (res *Response) NotFound(msg string) {
	if msg == "" {
		msg = "Not Found"
	}
	res.response(404, false, msg, nil, nil)
}

func (res *Response) InvalidRequest(msg string, errs any) {
	if msg == "" {
		msg = "Invalid Request"
	}
	res.response(400, false, msg, nil, errs)
}

func (res *Response) ServerErr(msg string) {
	if msg == "" {
		msg = "Internal Server Error"
	}
	res.response(500, false, msg, nil, nil)
}

func (res *Response) response(status int, success bool, msg string, data any, errs any) {
	body := map[string]any {
		"success": success,
		"message": msg,
	}
	if data != nil {
		body["data"] = data
	}
	if errs != nil {
		body["errors"] = errs
	}
	bodyJson, _ := json.Marshal(body)
	res.writer.Header().Set("Content-Type", "application/json") 
	res.writer.WriteHeader(status)
	res.writer.Write(bodyJson)
}
