package fixtures

import (
	"os"
	"path"
	"path/filepath"
	"runtime"
)

func SetTestConfigDir() {
	configDir := path.Join(srcDir(), "test", "fixtures", "configs")
	os.Setenv("DUOLINGO_CONFIG_DIR_PATH", configDir)
}

func srcDir() string {
	_, currentFile, _, _ := runtime.Caller(0)
	dir := filepath.Dir(currentFile)
	for {
		if pathExists(filepath.Join(dir, "go.mod")) {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	panic("go.mod not found; cannot determine project root")
}

func pathExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
