package restful

import (
	"context"
	"encoding/json"
	"io"

	"github.com/tidwall/gjson"
)

type RequestBuilder struct {
	request  *Request
	pathArgs map[string]string
	handler  any
}

func NewRequestBuilder(req *Request) *RequestBuilder {
	return &RequestBuilder{request: req}
}

func (b *RequestBuilder) GetRequestContext() context.Context {
	return b.request.base.Context()
}

func (b *RequestBuilder) SetRequestContext(ctx context.Context) {
	b.request.base = b.request.base.WithContext(ctx)
}

func (b *RequestBuilder) SetHandler(handler any) {
	b.handler = handler
}

func (b *RequestBuilder) SetPathArgs(args map[string]string) {
	b.pathArgs = args
}

func (b *RequestBuilder) Build() {
	b.buildPathArgsObject()
	b.buildQueriesObject()
	b.buildInputsObject()
	b.setRequestHandler()
}

func (b *RequestBuilder) buildPathArgsObject() {
	if js, jsErr := json.Marshal(b.pathArgs); jsErr == nil {
		b.request.pathArgs = gjson.ParseBytes(js)
	}
}

func (b *RequestBuilder) buildQueriesObject() {
	queriesMap := map[string][]string(b.request.URL().Query())
	reducedQueriesMap := make(map[string]any)
	for q := range queriesMap {
		if len(queriesMap[q]) == 1 {
			reducedQueriesMap[q] = queriesMap[q][0]
		}
		if len(queriesMap[q]) > 1 {
			reducedQueriesMap[q] = queriesMap[q]
		}
	}
	if js, jsErr := json.Marshal(reducedQueriesMap); jsErr == nil {
		b.request.queries = gjson.ParseBytes(js)
	}
}

func (b *RequestBuilder) buildInputsObject() {
	rawBody, readErr := io.ReadAll(b.request.base.Body)
	b.request.base.Body.Close()
	if readErr == nil {
		b.request.rawBody = rawBody
		b.request.inputs = gjson.ParseBytes(rawBody)
	}
}

func (b *RequestBuilder) setRequestHandler() {
	reqHandler, ok := b.handler.(func(*Request, *Response))
	if !ok {
		panic("invalid request handler")
	}
	b.request.handler = reqHandler
}
