package test_suites

import (
	user_repo "duolingo/repositories/user_repository/external"
	"duolingo/repositories/user_repository/external/services"
	"duolingo/repositories/user_repository/external/test/fixtures/data"

	"github.com/stretchr/testify/suite"
)

type UserServiceTestSuite struct {
	suite.Suite
	factory user_repo.UserRepoFactory
	repo    user_repo.UserRepository
	service services.UserService
}

func NewUserServiceTestSuite(
	factory user_repo.UserRepoFactory,
	repo user_repo.UserRepository,
	service services.UserService,
) *UserServiceTestSuite {
	return &UserServiceTestSuite{
		factory: factory,
		repo:    repo,
		service: service,
	}
}

func (s *UserServiceTestSuite) SetupTest() {
	s.repo.InsertManyUsers(data.TestUsers)
	s.repo.InsertManyUsers(data.TestUsersEmailUnverified)
}

func (s *UserServiceTestSuite) TearDownTest() {
	s.repo.DeleteUsersByIds(data.TestUserIds)
	s.repo.DeleteUsersByIds(data.TestUsersEmailUnverifiedIds)
}

func (s *UserServiceTestSuite) Test_CountDevicesForCampaign() {
	count, err := s.service.CountDevicesForCampaign(data.TestCampaignPrimary)
	s.Assert().NoError(err)
	s.Assert().Equal(uint64(len(data.TestDevices)), count)
}

func (s *UserServiceTestSuite) Test_GetDevicesForCampaign() {
	size := 2
	total := (len(data.TestDevices) + 1) / size // ceiling(len/size)
	for page := range total {
		devices, err := s.service.GetDevicesForCampaign(
			data.TestCampaignPrimary,
			uint64(page*size),
			uint64(size),
		)

		if !s.Assert().NoError(err) || !s.Assert().NotEmpty(devices) {
			return
		}

		for i := range size {
			if j := page*size + i; j < len(data.TestDevices) {
				s.Assert().Equal(*data.TestDevices[j], *devices[i])
			}
		}
	}
}
