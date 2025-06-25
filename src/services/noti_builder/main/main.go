package main

import (
	"context"
	cnst "duolingo/constants"
	ps "duolingo/libraries/pub_sub"
	container "duolingo/libraries/service_container"
	"duolingo/services/noti_builder/bootstrap"
	"duolingo/services/noti_builder/server"
	"duolingo/services/noti_builder/server/workloads"
)

func main() {
	bootstrap.Bootstrap()
	inputSubscriber := container.MustResolveAlias[ps.Subscriber](cnst.MesgInputSubscriber)
	notiPublisher := container.MustResolveAlias[ps.Publisher](cnst.PushNotiPublisher)
	tokenDistributor := container.MustResolve[*workloads.TokenBatchDistributor]()
	builder := server.NewNotiBuilder(inputSubscriber, notiPublisher, tokenDistributor)
	builder.Start(context.Background())
}
