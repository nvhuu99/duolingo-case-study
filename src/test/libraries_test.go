package test

import (
	"path/filepath"
	"runtime"
	"testing"

	"duolingo/lib/config-reader"
	mq "duolingo/lib/message-queue/driver/rabbitmq/test"
	wd "duolingo/lib/work-distributor/driver/redis/test"

	"github.com/stretchr/testify/suite"
)

var (
	_, caller, _, _ = runtime.Caller(0)
	dir = filepath.Dir(caller)
	conf = config.NewJsonReader(filepath.Join(dir, "..", "infra", "config"))
)

func TestMessageQueue(t *testing.T) {
	suite.Run(t, &mq.RabbitMQTestSuite{ 
		Host: conf.Get("mq.host", ""), 
		Port: conf.Get("mq.port", ""), 
		User: conf.Get("mq.user", ""), 
		Password: conf.Get("mq.pwd", ""),
	})
}

func TestWorkDistributor(t *testing.T) {
	suite.Run(t, &wd.RedisDistributorTestSuite{ 
		Host: conf.Get("redis.host", ""), 
		Port: conf.Get("redis.port", ""), 
	})
}