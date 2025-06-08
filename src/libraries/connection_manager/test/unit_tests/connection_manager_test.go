package unit_tests

import (
	"duolingo/libraries/connection_manager/test/test_suites"
	"testing"

	"github.com/stretchr/testify/suite"
)

func TestConnectionManager(t *testing.T) {
	suite.Run(t, &test_suites.ConnectionManagerTestSuite{})
}
