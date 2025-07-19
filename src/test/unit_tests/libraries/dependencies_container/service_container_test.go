package dependencies_container

import (
	"testing"

	"duolingo/libraries/dependencies_container/test/test_suites"

	"github.com/stretchr/testify/suite"
)

func TestServiceContainer(t *testing.T) {
	suite.Run(t, &test_suites.DependenciesContainerTestSuite{})
}
