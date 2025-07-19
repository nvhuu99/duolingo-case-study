package server

import (
	"context"
	"duolingo/libraries/restful"
	"duolingo/libraries/restful/router"
	"duolingo/libraries/restful/server/pipelines"
	"net/http"
	"strings"
	"time"
)

type Server struct {
	instance       *http.Server
	addr           string
	router         *router.Router
	pipelineGroups *restful.PipelineGroups
}

func NewServer(addr string) *Server {
	server := &Server{
		addr:           addr,
		router:         router.NewRouter(),
		pipelineGroups: restful.NewPipelineGroups(),
	}

	server.configServer()

	return server
}

func (server *Server) Serve(ctx context.Context) {
	server.startServer(ctx)
}

func (server *Server) Shutdown() {
	shutdownCtx, shutdownCancel := context.WithTimeout(
		context.Background(),
		5*time.Second,
	)
	defer shutdownCancel()
	server.instance.Shutdown(shutdownCtx)
}

func (server *Server) Addr() string {
	return server.addr
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

func (server *Server) startServer(ctx context.Context) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", server.handleRequest)
	server.instance = &http.Server{
		Addr:    server.addr,
		Handler: mux,
	}

	go func() {
		if err := server.instance.ListenAndServe(); err != nil {
			if err != http.ErrServerClosed {
				panic(err)
			}
		}
	}()

	<-ctx.Done()

	server.Shutdown()
}

func (server *Server) handleRequest(rw http.ResponseWriter, req *http.Request) {
	request := restful.NewRequest(req)
	response := restful.NewResponse(rw)
	defer server.handlePipelinePanic(response)
	server.pipelineGroups.ExecuteAll(request, response)
}

func (server *Server) handlePipelinePanic(response *restful.Response) {
	if r := recover(); r != nil {
		if !response.Sent() {
			response.ServerErr("Internal Server Error")
		}
	}
}
