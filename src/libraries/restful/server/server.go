package server

import (
	"duolingo/libraries/restful"
	"duolingo/libraries/restful/router"
	"duolingo/libraries/restful/server/pipelines"
	"net/http"
	"strings"
)

type Server struct {
	addr           string
	router         *router.Router
	pipelineGroups *restful.PipelineGroups
}

func NewServer(addr string) *Server {
	return &Server{
		addr:           addr,
		router:         router.NewRouter(),
		pipelineGroups: restful.NewPipelineGroups(),
	}
}

func (server *Server) Serve() {
	server.configServer()
	server.startServer()
}

func (server *Server) Get(path string, handler func(*restful.Request, *restful.Response)) {
	server.addRoute("GET", path, handler)
}

func (server *Server) Post(path string, handler func(*restful.Request, *restful.Response)) {
	server.addRoute("POST", path, handler)
}

func (server *Server) Put(path string, handler func(*restful.Request, *restful.Response)) {
	server.addRoute("PUT", path, handler)
}

func (server *Server) Delete(path string, handler func(*restful.Request, *restful.Response)) {
	server.addRoute("DELETE", path, handler)
}

func (server *Server) addRoute(
	method string,
	path string,
	handler func(*restful.Request, *restful.Response),
) {
	if strings.Trim(path, "/") == "" {
		panic("'/' is not a valid route")
	}
	if err := server.router.Add(method+"/"+path, handler); err != nil {
		panic(err)
	}
}

func (server *Server) configServer() {
	server.pipelineGroups.Push("requestFilter",
		&pipelines.HandlePreflightRequest{},
		&pipelines.ValidateCORS{},
		&pipelines.RouteRequestSetHandler{Router: server.router},
	)
	server.pipelineGroups.Push("handlingRequest",
		&pipelines.ExecuteRequestHandler{},
	)
	server.pipelineGroups.Push("requestHandled",
		&pipelines.FallbackNoContent{},
	)
}

func (server *Server) startServer() {
	http.HandleFunc("/", server.handleRequest)
	if err := http.ListenAndServe(server.addr, nil); err != nil {
		panic(err)
	}
}

func (server *Server) handleRequest(rw http.ResponseWriter, req *http.Request) {
	request := restful.NewRequest(req)
	response := restful.NewResponse(rw)
	defer server.panicHandler(response)
	server.pipelineGroups.ExecuteAll(request, response)
}

func (server *Server) panicHandler(response *restful.Response) {
	if r := recover(); r != nil {
		if !response.Sent() {
			response.ServerErr("Internal Server Error")
		}
	}
}
