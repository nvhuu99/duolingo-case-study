package test

import (
	"path/filepath"
	"runtime"
	"testing"

	"duolingo/lib/config-reader"
	mq "duolingo/lib/message-queue/driver/rabbitmq/test"

	"github.com/stretchr/testify/suite"
)

var (
	_, caller, _, _ = runtime.Caller(0)
	dir = filepath.Dir(caller)
	conf = config.NewJsonReader(filepath.Join(dir, "..", "infra", "config"))
	host = conf.Get("mq.host", "")
	port = conf.Get("mq.port", "")
	user = conf.Get("mq.user", "")
	pwd = conf.Get("mq.pwd", "")
)

func TestMessageQueue(t *testing.T) {
    suite.Run(t, &mq.RabbitMQManagerTestSuite{ Host: host, Port: port, User: user, Password: pwd })
    suite.Run(t, &mq.RabbitMQTopologyTestSuite{ Host: host, Port: port, User: user, Password: pwd })
}