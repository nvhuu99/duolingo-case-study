package dependencies

import (
	"context"
	"duolingo/libraries/dependencies_container"
	"duolingo/libraries/dependencies_provider"
	"sync"
)

var (
	registerOnce sync.Once
)

func RegisterDependencies(ctx context.Context) {
	registerOnce.Do(func() {
		dependencies_container.Init(ctx)
		dependencies_provider.Init()

		dependencies_provider.AddProvider(NewConfigReader(), "common")
		dependencies_provider.AddProvider(NewConnections(), "connections")
		dependencies_provider.AddProvider(NewMessageQueues(), "message_queues")
		dependencies_provider.AddProvider(NewUserRepo(), "user_repo")
		dependencies_provider.AddProvider(NewUserService(), "user_service")
		dependencies_provider.AddProvider(NewWorkDistributor(), "noti_builder")
		dependencies_provider.AddProvider(NewPushService(), "push_sender")
	})
}

func BootstrapDependencies(grps ...string) {
	dependencies_provider.BootstrapGroups(grps...)
}
