package common

import (
	"path/filepath"
	"runtime"
	"strings"

	"duolingo/common/container"
	"duolingo/lib/config"
)

var (
	serviceRootDir string // Root directory of the service.

	Config config.ConfigReader
)

func setServiceRoot() {
	_, filename, _, ok := runtime.Caller(2)
	if ok {
		serviceBootstrapDir := filepath.Dir(filename)
		serviceRootDir = filepath.Dir(serviceBootstrapDir)
	}
}

func SetupService() {
	setServiceRoot()
	Config = container.MakeConfigReader("json", Dir("config"))
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
