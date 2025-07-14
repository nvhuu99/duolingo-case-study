package router

import (
	"duolingo/libraries/restful/router"
	"testing"

	"github.com/stretchr/testify/suite"
)

func TestRouter(t *testing.T) {
	suite.Run(t, new(router.RouterTestSuite))
}
