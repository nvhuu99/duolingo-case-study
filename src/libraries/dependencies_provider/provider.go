package dependencies_provider

type DependenciesProvider interface {
	Bootstrap()
	Shutdown()
}
