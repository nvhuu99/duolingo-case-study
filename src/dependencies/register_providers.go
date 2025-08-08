package dependencies

import (
	"context"
	container "duolingo/libraries/dependencies_container"
	"duolingo/libraries/dependencies_provider"
	"sync"
)

var (
	registerOnce sync.Once
)

func registerDependencies(ctx context.Context) {
	registerOnce.Do(func() {
		container.Init(ctx)
		dependencies_provider.Init(ctx)
		dependencies_provider.AddProvider(NewConfigReader(), "common")
		dependencies_provider.AddProvider(NewEvents(), "event_manager")
		dependencies_provider.AddProvider(NewConnections(), "connections")
		dependencies_provider.AddProvider(NewMessageQueues(), "message_queues")
		dependencies_provider.AddProvider(NewUserRepo(), "user_repo")
		dependencies_provider.AddProvider(NewUserService(), "user_service")
		dependencies_provider.AddProvider(NewWorkDistributor(), "work_distributor")
		dependencies_provider.AddProvider(NewPushService(), "push_service")
	})
}

func Bootstrap(ctx context.Context, scope string, grps []string) {
	registerDependencies(ctx)
	dependencies_provider.BootstrapGroups(ctx, scope, grps)
}

func Shutdown(ctx context.Context) {
	dependencies_provider.ShutdownAll(ctx)
}
