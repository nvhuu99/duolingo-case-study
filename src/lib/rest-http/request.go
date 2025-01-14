package resthttp

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
)

type Request struct {
	req *http.Request
	
	parsed bool
	body []byte
	inputs map[string]any
}

func (request *Request) Header() http.Header {
	return request.req.Header
}

func (request *Request) Body() []byte {
	if !request.parsed {
		request.parseBody()
	}
	return request.body
}

func (request *Request) Path(key string) *RequestParam {
	val := request.req.PathValue(key)
	return &RequestParam{ value: val }
}

func (request *Request) Query(key string) *RequestParam {
	val := request.req.URL.Query().Get(key)
	return &RequestParam{ value: val }
}

func (request *Request) Input(pattern string) *RequestParam {
	if !request.parsed {
		request.parseBody()
	}

	var emptyValue = &RequestParam{ value: "" }

	if pattern == "" {
		return &RequestParam{ value: request.inputs }
	}

	iterator := request.inputs
	parts := strings.Split(pattern, ".")
	for i := 0; i < len(parts)-1; i++ {
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

	return &RequestParam{ value: iterator[key] }
}

func (request *Request) Has(key string) bool {
	return request.req.URL.Query().Has(key) || (request.req.PathValue(key) != "")
}

func (request *Request) parseBody() {
	request.body, _ = io.ReadAll(request.req.Body)
	request.req.Body.Close()
	err := json.Unmarshal(request.body, &request.inputs)
	if err != nil {
		request.inputs = make(map[string]any)
	}
	request.parsed = true
}