package main

import (
	"context"

	"duolingo/apps/message_input/server"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	server := server.NewMessageInputApiServer(ctx)
	server.Serve()
	defer server.Shutdown()
}
