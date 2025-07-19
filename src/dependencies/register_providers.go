package dependencies

import (
	"context"
	"duolingo/libraries/dependencies_container"
	"duolingo/libraries/dependencies_provider"
)

func RegisterDependencies(ctx context.Context) {
	dependencies_container.Init(ctx)
	dependencies_provider.Init()

	dependencies_provider.AddProvider(NewConfigReader(), "*")
	dependencies_provider.AddProvider(NewConnections(), "*")
	dependencies_provider.AddProvider(NewMessageQueues(), "*")
	dependencies_provider.AddProvider(NewUserRepo(), "*")
	dependencies_provider.AddProvider(NewUserService(), "*")
	dependencies_provider.AddProvider(NewWorkDistributor(), "noti_builder")
	dependencies_provider.AddProvider(NewPushService(), "push_sender")
}

func BootstrapDependencies(grps ...string) {
	dependencies_provider.BootstrapGroups(grps...)
}
