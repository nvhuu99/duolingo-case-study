package server

import (
	"context"
	"time"

	"duolingo/apps/message_input/server/handlers"
	"duolingo/dependencies"
	"duolingo/libraries/config_reader"
	container "duolingo/libraries/dependencies_container"
	restful "duolingo/libraries/restful/server"
	"duolingo/libraries/telemetry/otel_wrapper/log"
)

type MessageInputApiServer struct {
	ctx    context.Context
	server *restful.Server
	config config_reader.ConfigReader
	logger *log.Logger
}

func NewMessageInputApiServer(ctx context.Context) *MessageInputApiServer {
	dependencies.Bootstrap(ctx, "message_input", "", []string{
		"essentials",
		"connections",
		"message_queues",
		"pub_sub",
	})

	config := container.MustResolve[config_reader.ConfigReader]()
	server := restful.NewServer(config.Get("message_input", "server_address"))
	return &MessageInputApiServer{
		ctx:    ctx,
		server: server,
		config: config,
		logger: container.MustResolve[*log.Logger](),
	}
}

func (api *MessageInputApiServer) Addr() string {
	return api.server.Addr()
}

func (api *MessageInputApiServer) Serve() {
	api.server.Post("/api/v1/campaigns/{campaign}/message-input",
		handlers.NewMessageInputRequestHandler().Handle,
	)

	api.logger.Write(api.logger.
		Info("serving api").Namespace("message_input.api_server"))

	api.server.Serve(api.ctx)
}

func (api *MessageInputApiServer) Shutdown() {
	timeout := time.Duration(api.config.GetInt("message_input", "server_shutdown_wait"))
	ctx, cancel := context.WithTimeout(context.Background(), timeout*time.Second)
	defer cancel()

	dependencies.Shutdown(ctx)
	api.server.Shutdown(ctx)
}
