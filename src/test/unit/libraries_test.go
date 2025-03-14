package unit

import (
	"testing"

	mq "duolingo/lib/message-queue/driver/rabbitmq/test"
	wd "duolingo/lib/work-distributor/driver/redis/test"

	"github.com/stretchr/testify/suite"
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