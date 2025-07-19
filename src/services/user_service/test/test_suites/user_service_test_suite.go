package test_suites

import (
	container "duolingo/libraries/dependencies_container"
	user_repo "duolingo/repositories/user_repository/external"
	"duolingo/services/user_service"
	"duolingo/test/fixtures/data"

	"github.com/stretchr/testify/suite"
)

type UserServiceTestSuite struct {
	suite.Suite
	repo    user_repo.UserRepository
	service *user_service.UserService
}

func NewUserServiceTestSuite() *UserServiceTestSuite {
	return &UserServiceTestSuite{
		repo:    container.MustResolve[user_repo.UserRepository](),
		service: container.MustResolve[*user_service.UserService](),
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
	s.Assert().Equal(int64(len(data.TestDevices)), count)
}

func (s *UserServiceTestSuite) Test_GetDevicesForCampaign() {
	size := 2
	total := (len(data.TestDevices) + 1) / size // ceiling(len/size)
	for page := range total {
		devices, err := s.service.GetDevicesForCampaign(
			data.TestCampaignPrimary,
			int64(page*size),
			int64(size),
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
