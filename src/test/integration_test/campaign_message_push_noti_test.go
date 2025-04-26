// Warning: DO NOT run the test on production environment, data lost could be happen.
// You should only run it on test environments.
package integration

import (
	"path/filepath"
	"testing"

	config "duolingo/lib/config_reader/driver/reader/json"
	its "duolingo/test/suite/integration"

	"github.com/stretchr/testify/suite"
)

func TestCampaignMessagePushNotification(t *testing.T) {
	configDir, _ := filepath.Abs(filepath.Join("..", "..", "config"))
	servicesDir, _ := filepath.Abs(filepath.Join("..", "..", "service"))
	suite.Run(t, &its.CampaignMessagePushNotiTestSuite{
		ServiceDir:   servicesDir,
		ConfigReader: config.NewJsonReader(configDir),
		Campaign:     "test_campaign",
	})
}
