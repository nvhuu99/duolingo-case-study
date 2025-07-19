package main

import (
	"context"
	"log"

	"duolingo/apps/message_input/server"
	"duolingo/dependencies"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	dependencies.RegisterDependencies(ctx)
	dependencies.BootstrapDependencies("", []string{
		"common",
		"connections",
		"message_queues",
	})

	server := server.NewMessageInputApiServer()
	server.RegisterRoutes()

	log.Println("serving message input api")
	server.Serve(ctx)
}
