package restful

import (
	"duolingo/libraries/restful/router"
	"net/http"
	"strings"
)

type Server struct {
	addr           string
	router         *router.Router
	pipelineGroups *PipelineGroups
}

func NewServer(addr string) *Server {
	return &Server{
		addr:           addr,
		router:         router.NewRouter(),
		pipelineGroups: NewPipelineGroups(),
	}
}

func (server *Server) Serve() {
	server.configServer()
	server.startServer()
}

func (server *Server) Get(path string, handler func(*Request, *Response)) {
	server.addRoute("GET", path, handler)
}

func (server *Server) Post(path string, handler func(*Request, *Response)) {
	server.addRoute("POST", path, handler)
}

func (server *Server) Put(path string, handler func(*Request, *Response)) {
	server.addRoute("PUT", path, handler)
}

func (server *Server) Delete(path string, handler func(*Request, *Response)) {
	server.addRoute("DELETE", path, handler)
}

func (server *Server) addRoute(method string, path string, handler func(*Request, *Response)) {
	if strings.Trim(path, "/") == "" {
		panic("'/' is not a valid route")
	}
	if err := server.router.Add(method+"/"+path, handler); err != nil {
		panic(err)
	}
}

func (server *Server) configServer() {
	server.pipelineGroups.Push("requestFilter",
		&handleRequestPreflight{},
		&validateCORS{},
		&routeRequest{Router: server.router},
	)
	server.pipelineGroups.Push("handlingRequest",
		&executeRequestHandler{},
	)
	server.pipelineGroups.Push("requestHandled",
		&fallbackNoContent{},
	)
}

func (server *Server) startServer() {
	http.HandleFunc("/", server.handleRequest)
	if err := http.ListenAndServe(server.addr, nil); err != nil {
		panic(err)
	}
}

func (server *Server) handleRequest(rw http.ResponseWriter, req *http.Request) {
	request := &Request{base: req}
	response := &Response{base: rw}
	defer server.panicHandler(response)
	server.pipelineGroups.ExecuteAll(request, response)
}

func (server *Server) panicHandler(response *Response) {
	if r := recover(); r != nil {
		if !response.sent {
			response.ServerErr("Internal Server Error")
		}
	}
}
