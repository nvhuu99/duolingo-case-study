package main

import (
	"log"

	"duolingo/lib/log/grpc_service"
	sv "duolingo/lib/service_container"
	"duolingo/service/log_service/bootstrap"
)

var (
	container *sv.ServiceContainer
	server    *grpc_service.LoggerGRPCServer
)

func main() {
	bootstrap.Run()

	container = sv.GetContainer()
	server = container.Resolve("server.grpc_logger_server").(*grpc_service.LoggerGRPCServer)

	log.Println("log server started")

	err := server.Listen()

	if err != nil {
		panic(err)
	}
}
