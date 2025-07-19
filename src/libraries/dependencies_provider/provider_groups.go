package dependencies_provider

import (
	"sync"
)

var (
	providerGroups  *DependenciesProviderGroups
	ensureSingleton sync.Once
)

type DependenciesProviderGroups struct {
	groups       map[string][]int
	providers    map[int]DependenciesProvider
	bootedGroups map[string]bool
}

func Init() *DependenciesProviderGroups {
	ensureSingleton.Do(func() {
		providerGroups = &DependenciesProviderGroups{
			groups:       make(map[string][]int),
			providers:    make(map[int]DependenciesProvider),
			bootedGroups: make(map[string]bool),
		}
	})
	return providerGroups
}

func AddProvider(provider DependenciesProvider, grps ...string) {
	idx := len(providerGroups.providers)
	providerGroups.providers[idx] = provider
	for _, g := range grps {
		providerGroups.groups[g] = append(providerGroups.groups[g], idx)
	}
}

func BootstrapGroups(grps ...string) {
	for _, g := range grps {
		if providerGroups.bootedGroups[g] {
			continue
		}
		if indexes, exists := providerGroups.groups[g]; exists {
			for _, i := range indexes {
				provider := providerGroups.providers[i]
				provider.Bootstrap()
			}
		}
		providerGroups.bootedGroups[g] = true
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
