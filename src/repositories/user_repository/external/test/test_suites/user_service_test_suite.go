package test_suites

import (
	"duolingo/models"
	user_repo "duolingo/repositories/user_repository/external"
	"duolingo/repositories/user_repository/external/services"
	"math/rand"
	"strings"
	"time"

	"github.com/stretchr/testify/suite"
)

type UserServiceTestSuite struct {
	suite.Suite
	factory      user_repo.UserRepoFactory
	repo         user_repo.UserRepository
	service      services.UserService
	testUsersMap map[string]*models.User
	testUserIds  []string
}

func NewUserServiceTestSuite(factory user_repo.UserRepoFactory) *UserServiceTestSuite {
	return &UserServiceTestSuite{
		factory: factory,
		service: factory.MakeUserService(),
		repo:    factory.MakeUserRepo(),
	}
}

func (s *UserServiceTestSuite) SetupTest() {
	insertedUsrs, err := s.insertFakeUsers()
	if err != nil {
		s.FailNow(err.Error())
	}
	s.testUsersMap = make(map[string]*models.User)
	s.testUserIds = make([]string, len(insertedUsrs))
	for i := range insertedUsrs {
		s.testUsersMap[insertedUsrs[i].Id] = insertedUsrs[i]
		s.testUserIds[i] = insertedUsrs[i].Id
	}
}

func (s *UserServiceTestSuite) TearDownTest() {
	s.repo.DeleteUsersByIds(s.testUserIds)
}

func (s *UserServiceTestSuite) Test_CountDevicesForCampaign() {
	count, err := s.service.CountDevicesForCampaign(testOnlyCampaign)
	s.Assert().NoError(err)
	s.Assert().Equal(uint64(len(s.testUserIds)*len(testDevices)), count)
}

func (s *UserServiceTestSuite) Test_GetDevicesForCampaign() {
	devices, err := s.service.GetDevicesForCampaign(testOnlyCampaign, 0, 100)
	s.Assert().NoError(err)

	if s.Assert().Equal(len(s.testUserIds)*len(testDevices), len(devices)) {
		for _, device := range devices {
			s.Assert().Equal(device.Platform, "fake_platform")
			s.Assert().True(strings.HasPrefix(device.Token, "fake_token_"))
		}
	}
}

func (s *UserServiceTestSuite) insertFakeUsers() ([]*models.User, error) {
	n := rand.Intn(5) + 1
	usrs := make([]*models.User, n)
	for i := range n {
		usrs[i] = &models.User{
			Campaigns:       []string{testOnlyCampaign},
			Devices:         testDevices,
			EmailVerifiedAt: time.Now().Add(-1 * time.Hour),
		}
	}
	return s.repo.InsertManyUsers(usrs)
}
