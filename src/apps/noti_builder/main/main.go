package main

import (
	"context"

	"duolingo/apps/noti_builder/server"
	"duolingo/dependencies"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	dependencies.Bootstrap(ctx, "noti_builder", "", []string{
		"essentials",
		"connections",
		"message_queues",
		"pub_sub",
		"task_queues",
		"user_repo",
		"user_service",
		"work_distributor",
	})

	builder := server.NewNotiBuilder()
	builder.Start(ctx)
}
