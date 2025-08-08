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

	dependencies.Bootstrap(ctx, "", []string{
		"common",
		"connections",
		"message_queues",
		"push_service",
	})

	sender := server.NewSender()
	log.Println("running push notification sender")
	sender.Start(ctx)
}
