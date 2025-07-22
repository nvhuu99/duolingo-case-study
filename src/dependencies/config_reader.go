package dependencies

import (
	"context"
	"log"
	"os"

	"duolingo/libraries/config_reader"
	container "duolingo/libraries/dependencies_container"
)

type ConfigReader struct {
	configDir string
}

func NewConfigReader() *ConfigReader {
	fromEnv := os.Getenv("DUOLINGO_CONFIG_DIR_PATH")
	if fromEnv == "" {
		panic("environment variable DUOLINGO_CONFIG_DIR_PATH is not set")
	} else {
		log.Printf("%v is use as the configurations files directory\n", fromEnv)
	}
	return &ConfigReader{configDir: fromEnv}
}

func (c *ConfigReader) Bootstrap(scope string) {
	container.BindSingleton[config_reader.ConfigReader](func(ctx context.Context) any {
		return config_reader.
			NewJsonConfigReader().
			LoadFromLocalFiles(c.configDir)
	})
}

func (c *ConfigReader) Shutdown() {
}
