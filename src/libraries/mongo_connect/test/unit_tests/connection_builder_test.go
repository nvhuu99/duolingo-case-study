package unit_tests

import (
	"duolingo/libraries/mongo_connect/test/test_suites"
	"testing"

	"github.com/stretchr/testify/suite"
)

func TestUserRepositoryBuilder(t *testing.T) {
	suite.Run(t, &test_suites.ConnectionBuilderTestSuite{})
}
