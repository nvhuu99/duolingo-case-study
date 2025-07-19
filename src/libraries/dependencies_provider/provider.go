package dependencies_provider

type DependenciesProvider interface {
	Bootstrap(scope string)
	Shutdown()
}
