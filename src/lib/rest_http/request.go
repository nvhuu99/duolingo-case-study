package rest_http

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
)

type Request struct {
	id        string
	req     *http.Request
	rawBody []byte
	inputs  map[string]any

	Timestamp time.Time
}

func ParseRequest(req *http.Request) *Request {
	request := &Request{
		id:  uuid.New().String(),
		req: req,
		Timestamp: time.Now(),
	}
	request.parseBody()
	return request
}

func (request *Request) Instance() *http.Request {
	return request.req
}

func (request *Request) Id() string {
	return request.id
}

func (request *Request) Method() string {
	return request.req.Method
}

func (request *Request) URL() *url.URL {
	return request.req.URL
}

func (request *Request) Header() http.Header {
	return request.req.Header
}

func (request *Request) Body() []byte {
	return request.rawBody
}

func (request *Request) Path(key string) *Value {
	return &Value{request.req.PathValue(key)}
}

func (request *Request) Query(key string) *Value {
	if key == "*" {
		queryParams := make(map[string]string)
		for k, v := range request.URL().Query() {
			queryParams[k] = v[0]
		}

		return &Value{ queryParams }
	}

	return &Value{request.req.URL.Query().Get(key)}
}

func (request *Request) Input(pattern string) *Value {
	var emptyValue = &Value{""}

	if pattern == "*" {
		return &Value{request.inputs}
	}

	iterator := request.inputs
	parts := strings.Split(pattern, ".")
	for i := range len(parts) - 1 {
		p := parts[i]
		if _, exists := iterator[p]; !exists {
			return emptyValue
		}
		next, ok := iterator[p].(map[string]interface{})
		if !ok {
			return emptyValue
		}
		iterator = next
	}

	key := parts[len(parts)-1]
	if _, exists := iterator[key]; !exists {
		return emptyValue
	}

	return &Value{iterator[key]}
}

func (request *Request) parseBody() {
	request.rawBody, _ = io.ReadAll(request.req.Body)
	request.req.Body.Close()
	err := json.Unmarshal(request.rawBody, &request.inputs)
	if err != nil {
		request.inputs = make(map[string]any)
	}
}
