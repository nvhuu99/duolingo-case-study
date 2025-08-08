package dependencies_provider

import "context"

type DependenciesProvider interface {
	Bootstrap(ctx context.Context, scope string)
	Shutdown(ctx context.Context)
}
