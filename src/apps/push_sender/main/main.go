package main

import (
	"context"

	"duolingo/apps/push_sender/server"
	"duolingo/dependencies"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	dependencies.Bootstrap(ctx, "push_sender", "", []string{
		"essentials",
		"connections",
		"task_queues",
		"push_service",
	})

	sender := server.NewSender()
	sender.Start(ctx)
}
