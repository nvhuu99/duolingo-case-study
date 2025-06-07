package unit_tests

import (
	"duolingo/repositories/campaign_message/user_repository/drivers/mongodb/test/test_suites"
	"testing"

	"github.com/stretchr/testify/suite"
)

func TestUserRepositoryBuilder(t *testing.T) {
	suite.Run(t, new(test_suites.UserRepoBuilderTestSuite))
}
