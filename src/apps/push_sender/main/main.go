package main

import (
	"context"
	"log"

	"duolingo/apps/push_sender/server"
	"duolingo/dependencies"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	dependencies.RegisterDependencies(ctx)
	dependencies.BootstrapDependencies("push_sender")

	sender := server.NewSender()
	log.Println("running push notification sender")
	sender.Start(ctx)
}
