package test

import (
	"testing"

	mq "duolingo/lib/message-queue/driver/rabbitmq/test"

	"github.com/stretchr/testify/suite"
)

func TestMessageQueue(t *testing.T) {
    suite.Run(t, new(mq.RabbitMQManagerTestSuite))
}