package firebase

import (
	"testing"

	"duolingo/libraries/push_notification/drivers/firebase/test/test_suites"

	"github.com/stretchr/testify/suite"
)

func TestMessageBuilder(t *testing.T) {
	suite.Run(t, &test_suites.MessageBuilderTestSuite{})
}
