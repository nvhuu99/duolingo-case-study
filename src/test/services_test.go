package test

import (
	"testing"

	mqInputApi "duolingo/services/message-input-api/test"

	"github.com/stretchr/testify/suite"
)


func TestMessageInputApi(t *testing.T) {
	suite.Run(t, &mqInputApi.MessageInputApiTestSuite{ 
		Host: conf.Get("mq.host", ""), 
		Port: conf.Get("mq.port", ""), 
		User: conf.Get("mq.user", ""), 
		Password: conf.Get("mq.pwd", ""),
	})
}
