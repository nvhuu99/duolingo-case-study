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
