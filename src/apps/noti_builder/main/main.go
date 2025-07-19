package main

import (
	"context"
	"log"

	"duolingo/apps/noti_builder/server"
	"duolingo/dependencies"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	dependencies.RegisterDependencies(ctx)

	builder := server.NewNotiBuilder()
	log.Println("running notification builder")
	builder.Start(ctx)
}
