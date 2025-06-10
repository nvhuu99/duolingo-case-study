package work_distributor

import (
	"testing"

	"duolingo/libraries/work_distributor/test/test_suites"

	"github.com/stretchr/testify/suite"
)

func TestAssignment(t *testing.T) {
	suite.Run(t, test_suites.NewAssignmentTestSuite())
}
