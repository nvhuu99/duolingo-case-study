package restful

import (
	"net/http"
	"net/url"

	"github.com/tidwall/gjson"
)

type Request struct {
	base     *http.Request
	rawBody  []byte
	queries  gjson.Result
	pathArgs gjson.Result
	inputs   gjson.Result
	handler  func(*Request, *Response)
}

func (request *Request) Method() string {
	return request.base.Method
}

func (request *Request) URL() *url.URL {
	return request.base.URL
}

func (request *Request) Header() http.Header {
	return request.base.Header
}

func (request *Request) PathArg(key string) gjson.Result {
	return request.pathArgs.Get(key)
}

func (request *Request) Query(key string) gjson.Result {
	return request.queries.Get(key)
}

func (request *Request) HasQueries(keys ...string) bool {
	for i := range keys {
		if !request.queries.Get(keys[i]).Exists() {
			return false
		}
	}
	return true
}

func (request *Request) Input(path string) gjson.Result {
	return request.inputs.Get(path)
}

func (request *Request) RawBody() []byte {
	return request.rawBody
}
