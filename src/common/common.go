package common

import (
	"context"
	"path/filepath"
	"runtime"
	"strings"

	sv "duolingo/lib/service-container"
	"duolingo/lib/config-reader"
)

var (
	serviceRootDir string
	serviceContext context.Context
	serviceContextCancel context.CancelFunc
	serviceContainer *sv.ServiceContainer
)

func Container() *sv.ServiceContainer {
	if serviceContainer == nil {
		serviceContainer = sv.NewContainer()
	}
	return serviceContainer
}

func SetupService() {
	container := Container()
	// Set service root
	_, filename, _, _ := runtime.Caller(2)
	serviceBootstrapDir := filepath.Dir(filename)
	serviceRootDir = filepath.Dir(serviceBootstrapDir)
	
	// Set service context
	serviceContext, serviceContextCancel = context.WithCancel(context.Background())
	
	// Services binding
	container.BindSingleton("config", func() any {
		return config.NewJsonReader(Dir("config"))
	})
	
	container.BindSingleton("config.infra", func() any {
		return config.NewJsonReader(Dir("..", "..", "infra", "config"))
	})
}

func ServiceContext() (context.Context, context.CancelFunc) {
	return serviceContext, serviceContextCancel
}

// Dir constructs an absolute path by appending the provided parts to the service's root directory.
// The root directory is determined dynamically based on the location of the `common` package.
func Dir(parts ...string) string {
	if len(parts) == 1 {
		parts = strings.Split(parts[0], "/")
	}
	// Prepend the root directory to the provided parts and join them into a path.
	parts = append([]string{serviceRootDir}, parts...)
	return filepath.Join(parts...)
}
