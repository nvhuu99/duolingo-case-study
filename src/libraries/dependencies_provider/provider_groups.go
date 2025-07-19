package dependencies_provider

import (
	"sync"
)

var (
	providerGroups  *DependenciesProviderGroups
	ensureSingleton sync.Once
)

type DependenciesProviderGroups struct {
	groups    map[string][]int
	providers map[int]DependenciesProvider
}

func Init() *DependenciesProviderGroups {
	ensureSingleton.Do(func() {
		providerGroups = &DependenciesProviderGroups{
			groups:    make(map[string][]int),
			providers: make(map[int]DependenciesProvider),
		}
	})
	return providerGroups
}

func AddProvider(provider DependenciesProvider, grps ...string) {
	idx := len(providerGroups.providers) + 1
	providerGroups.providers[idx] = provider
	for _, g := range grps {
		providerGroups.groups[g] = append(providerGroups.groups[g], idx)
	}
}

func BootstrapGroups(grps ...string) {
	grps = append(grps, "*")
	for _, g := range grps {
		if indexes, exists := providerGroups.groups[g]; exists {
			for i := range indexes {
				provider := providerGroups.providers[i]
				provider.Bootstrap()
			}
		}
	}
}

func ShutdownGroups(grps ...string) {
	for _, g := range grps {
		if indexes, exists := providerGroups.groups[g]; exists {
			for i := range indexes {
				provider := providerGroups.providers[i]
				provider.Shutdown()
			}
		}
	}
}
