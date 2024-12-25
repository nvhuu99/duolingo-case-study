package common

import (
	"duolingo/lib/config"
	"path/filepath"
	"runtime"
)

var (
	rootDir      string           // Root directory of the service.
	configReader config.ConfigReader // Singleton instance of the configuration reader.
)

// Config initializes and returns a singleton instance of the configuration reader.
// The reader uses the "config" directory relative to the service's root directory.
func Config() config.ConfigReader {
	if configReader == nil {
		configReader = config.NewJsonReader(Dir("config"))
	}
	return configReader
}

// Dir constructs an absolute path by appending the provided parts to the service's root directory.
// The root directory is determined dynamically based on the location of the `common` package.
func Dir(parts ...string) string {
	if rootDir == "" {
		// Retrieve the directory of the current file.
		_, filename, _, ok := runtime.Caller(0)
		if ok {
			commonDir := filepath.Dir(filename) // Directory of the current file.
			rootDir = filepath.Dir(commonDir)  // Root directory of the service.
		}
	}

	// Prepend the root directory to the provided parts and join them into a path.
	parts = append([]string{rootDir}, parts...)
	return filepath.Join(parts...)
}
