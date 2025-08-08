package server

import (
	"context"
	"log"
	"time"

	"duolingo/apps/message_input/server/handlers"
	"duolingo/dependencies"
	"duolingo/libraries/config_reader"
	container "duolingo/libraries/dependencies_container"
	restful "duolingo/libraries/restful/server"
)

type MessageInputApiServer struct {
	ctx context.Context
	server *restful.Server
	config config_reader.ConfigReader
}

func NewMessageInputApiServer(ctx context.Context) *MessageInputApiServer {
	dependencies.Bootstrap(ctx, "", []string{
		"common",
		"event_manager",
		"connections",
		"message_queues",
	})

	config := container.MustResolve[config_reader.ConfigReader]()
	server := restful.NewServer(config.Get("message_input", "server_address"))
	return &MessageInputApiServer{
		ctx: ctx,
		server: server,
		config: config,
	}
}

func (api *MessageInputApiServer) Addr() string {
	return api.server.Addr()
}

func (api *MessageInputApiServer) Serve() {
	api.server.Post("/api/v1/campaigns/{campaign}/message-input",
		handlers.NewMessageInputRequestHandler().Handle,
	)

	log.Println("serving api")

	api.server.Serve(api.ctx)
}

func (api *MessageInputApiServer) Shutdown() {
	timeout := time.Duration(api.config.GetInt("message_input", "server_shutdown_wait"))
	ctx, cancel := context.WithTimeout(context.Background(), timeout * time.Second)
	defer cancel()

	dependencies.Shutdown(ctx)
	api.server.Shutdown(ctx)
}
