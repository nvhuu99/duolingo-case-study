package main

import (
	"context"

	"duolingo/apps/noti_builder/server"
	"duolingo/dependencies"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	dependencies.Bootstrap(ctx, "", []string{
		"common",
		"connections",
		"message_queues",
		"user_repo",
		"user_service",
		"work_distributor",
	})

	builder := server.NewNotiBuilder()
	builder.Start(ctx)
}
