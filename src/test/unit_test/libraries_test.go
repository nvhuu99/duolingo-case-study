package unit

import (
	"path/filepath"
	"testing"

	config "duolingo/lib/config_reader/driver/reader/json"
	mq "duolingo/lib/message_queue/driver/rabbitmq/test"
	wd "duolingo/lib/work_distributor/driver/redis/test"

	"github.com/stretchr/testify/suite"
)

var (
	dir, _ = filepath.Abs(filepath.Join("..", "..", "config"))
	conf   = config.NewJsonReader(dir)
)

func TestMessageQueue(t *testing.T) {
	suite.Run(t, &mq.RabbitMQTestSuite{
		Host:     conf.Get("mq.host", ""),
		Port:     conf.Get("mq.port", ""),
		User:     conf.Get("mq.user", ""),
		Password: conf.Get("mq.pwd", ""),
	})
}

func TestWorkDistributor(t *testing.T) {
	suite.Run(t, &wd.RedisDistributorTestSuite{
		Host: conf.Get("redis.host", ""),
		Port: conf.Get("redis.port", ""),
	})
}
