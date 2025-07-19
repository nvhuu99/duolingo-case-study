package server

import (
	"context"

	"duolingo/apps/message_input/server/handlers"
	"duolingo/libraries/config_reader"
	container "duolingo/libraries/dependencies_container"
	restful "duolingo/libraries/restful/server"
)

type MessageInputApiServer struct {
	server *restful.Server
}

func NewMessageInputApiServer() *MessageInputApiServer {
	config := container.MustResolve[config_reader.ConfigReader]()
	server := restful.NewServer(config.Get("message_input", "server_address"))
	return &MessageInputApiServer{
		server: server,
	}
}

func (api *MessageInputApiServer) Addr() string {
	return api.server.Addr()
}

func (api *MessageInputApiServer) Serve(ctx context.Context) {
	api.server.Serve(ctx)
}

func (api *MessageInputApiServer) Shutdown() {
	api.server.Shutdown()
}

func (api *MessageInputApiServer) RegisterRoutes() {
	api.server.Post("/api/v1/campaigns/{campaign}/message-input",
		handlers.NewMessageInputRequestHandler().Handle,
	)
}
