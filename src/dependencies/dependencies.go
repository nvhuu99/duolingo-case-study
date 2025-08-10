package dependencies

import (
	"context"
	"duolingo/dependencies/providers"
	"duolingo/libraries/dependencies_container"
	"duolingo/libraries/dependencies_provider"
	"os"
	"sync"
)

var (
	registerOnce sync.Once
)

func Bootstrap(ctx context.Context, app string, scope string, grps []string) {
	registerOnce.Do(func() {
		os.Setenv("DUOLINGO_APP_NAME", app)

		dependencies_container.Init(ctx)
		dependencies_provider.Init(ctx)

		dependencies_provider.AddProvider(&providers.ConfigReaderProvider{}, "essentials")
		dependencies_provider.AddProvider(&providers.TraceManagerProvider{}, "essentials")
		dependencies_provider.AddProvider(&providers.EventManagerProvider{}, "essentials")

		dependencies_provider.AddProvider(&providers.ConnectionsProvider{}, "connections")
		dependencies_provider.AddProvider(&providers.MessageQueuesProvider{}, "message_queues")
		dependencies_provider.AddProvider(&providers.PubSubProvider{}, "message_queues", "pub_sub")
		dependencies_provider.AddProvider(&providers.TaskQueueProvider{}, "message_queues", "task_queues")
		dependencies_provider.AddProvider(&providers.UserRepoProvider{}, "user_repo")
		dependencies_provider.AddProvider(&providers.UserServiceProvider{}, "user_service")
		dependencies_provider.AddProvider(&providers.WorkDistributorProvider{}, "work_distributor")
		dependencies_provider.AddProvider(&providers.PushServiceProvider{}, "push_service")

		dependencies_provider.BootstrapGroups(ctx, scope, grps)
	})
}

func Shutdown(ctx context.Context) {
	dependencies_provider.ShutdownAll(ctx)
}
