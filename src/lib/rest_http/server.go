package rest_http

import (
	"errors"
	"fmt"
	"net/http"
	"time"
)

type Server struct {
	addr        string
	router      *Router
	middlewares *MiddlewareManager
}

func NewServer(addr string) *Server {
	server := new(Server)
	server.addr = addr
	server.middlewares = NewMiddlewareManager()
	server.router = NewRouter()
	return server
}

func (server *Server) WithMiddlewares(group string, middlwares ...Handler) *Server {
	for _, handler := range middlwares {
		server.middlewares.Push(group, handler)
	}
	return server
}

func (server *Server) Router() *Router {
	return server.router
}

func (server *Server) Serve() {
	server.configServer()
	server.startServer()
}

func (server *Server) SendResponse(request *Request, response *Response) {
	if response.ResponseSent {
		return
	}

	body := response.GetBody()
	response.ResponseBodySize = len(body)
	response.ResponseTimeMs = int(time.Since(request.Timestamp).Milliseconds())
	if response.Status == 0 {
		response.Status = STATUS_OK
	}
	response.ResponseSent = true

	response.rw.Header().Set("Content-Type", "application/json")
	response.rw.WriteHeader(response.Status)
	response.rw.Write(body)
}

func (server *Server) configServer() {
	// middlewares
	server.middlewares.Push("request", &RouteRequest{server: server})
	// set server request handler
	http.HandleFunc("/", func(rw http.ResponseWriter, req *http.Request) {
		request := ParseRequest(req)
		response := NewResponse(rw)

		defer server.panicHandler(request, response)

		server.middlewares.Handle("request", request, response)
		server.SendResponse(request, response)
		server.middlewares.Handle("response", request, response)
	})
}

func (server *Server) startServer() {
	err := http.ListenAndServe(server.addr, nil)
	if err != nil {
		panic(fmt.Errorf("%v - %w", ErrMessages[ERR_SERVER_PANIC], err))
	}
}

func (server *Server) panicHandler(request *Request, response *Response) {
	if r := recover(); r != nil {
		if err, ok := r.(error); ok {
			response.Errors = []error{
				fmt.Errorf("%v - %w", ErrMessages[ERR_SERVER_PANIC], err),
			}
		} else {
			response.Errors = []error{
				errors.New(ErrMessages[ERR_SERVER_PANIC]),
			}
		}
		server.SendResponse(request, response.ServerErr("", nil))
	}
}
