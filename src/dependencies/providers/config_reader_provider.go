package providers

import (
	"context"
	"log"
	"os"

	"duolingo/libraries/config_reader"
	container "duolingo/libraries/dependencies_container"
)

type ConfigReaderProvider struct {
}

func (c *ConfigReaderProvider) Bootstrap(bootstrapCtx context.Context, scope string) {
	configDir := os.Getenv("DUOLINGO_CONFIG_DIR_PATH")
	if configDir == "" {
		panic("environment variable DUOLINGO_CONFIG_DIR_PATH is not set")
	}
	log.Printf("use %v as configuration files directory\n", configDir)

	container.BindSingleton[config_reader.ConfigReader](func(ctx context.Context) any {
		return config_reader.
			NewJsonConfigReader().
			LoadFromLocalFiles(configDir)
	})
}

func (c *ConfigReaderProvider) Shutdown(shutdownCtx context.Context) {
}
