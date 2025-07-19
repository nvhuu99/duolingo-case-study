package dependencies

import (
	"context"
	"duolingo/libraries/config_reader"
	container "duolingo/libraries/dependencies_container"
	"os"
)

type ConfigReader struct {
	configDir string
}

func NewConfigReader() *ConfigReader {
	configDir := ".tmp/configs"
	if fromEnv := os.Getenv("DUOLINGO_CONFIG_DIR_PATH"); fromEnv != "" {
		configDir = fromEnv
	}
	return &ConfigReader{configDir}
}

func (c *ConfigReader) Bootstrap() {
	container.BindSingleton[config_reader.ConfigReader](func(ctx context.Context) any {
		return config_reader.
			NewJsonConfigReader().
			SetLocalFileDir(c.configDir).
			AddLocalFile("rabbitmq", "rabbitmq.json").
			AddLocalFile("mongodb", "mongodb.json").
			AddLocalFile("redis", "redis.json").
			AddLocalFile("message_input", "message_input.json").
			AddLocalFile("noti_builder", "noti_builder.json").
			AddLocalFile("push_sender", "push_sender.json").
			AddLocalFile("firebase", "firebase.json").
			AddLocalFile("work_distributor", "work_distributor.json")
	})
}

func (c *ConfigReader) Shutdown() {
}
