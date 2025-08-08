package dependencies_provider

import (
	"context"
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

func Init(ctx context.Context) *DependenciesProviderGroups {
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

func BootstrapGroups(ctx context.Context, scope string, grps []string) {
	for _, g := range grps {
		if providerGroups.bootedGroups[g] {
			continue
		}
		if indexes, exists := providerGroups.groups[g]; exists {
			for _, i := range indexes {
				provider := providerGroups.providers[i]
				provider.Bootstrap(ctx, scope)
			}
		}
		providerGroups.bootedGroups[g] = true
	}
}

func ShutdownAll(ctx context.Context) {
	for g := range providerGroups.bootedGroups {
		if indexes, exists := providerGroups.groups[g]; exists {
			for i := range indexes {
				provider := providerGroups.providers[i]
				provider.Shutdown(ctx)
			}
		}
	}
}
