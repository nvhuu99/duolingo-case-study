package buffer

import (
	"duolingo/libraries/buffer/test/test_suites"
	"testing"

	"github.com/stretchr/testify/suite"
)

func TestBufferGroup(t *testing.T) {
	suite.Run(t, &test_suites.BufferGroupTestSuite{})
}
