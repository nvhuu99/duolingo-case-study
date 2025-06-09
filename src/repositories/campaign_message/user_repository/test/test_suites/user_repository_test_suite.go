package test_suites

import (
	user_repository "duolingo/repositories/campaign_message/user_repository"
	"duolingo/repositories/campaign_message/user_repository/models"
	"math/rand"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
)

var (
	testOnlyCampaign = "testOnlyCampaign"
	testDeviceTokens = []string{"fake_token_1", "fake_token_2"}
)

type UserRepositoryTestSuite struct {
	suite.Suite
	repo         user_repository.UserRepository
	testUsersMap map[string]*models.User
	testUserIds  []string
}

func NewUserRepositoryTestSuite(repo user_repository.UserRepository) *UserRepositoryTestSuite {
	return &UserRepositoryTestSuite{
		repo: repo,
	}
}

func (s *UserRepositoryTestSuite) SetupTest() {
	insertedUsrs, err := s.insertFakeUsers()
	if err != nil {
		s.T().Log(err)
		panic("fail to setup test")
	}
	s.testUsersMap = make(map[string]*models.User)
	s.testUserIds = make([]string, len(insertedUsrs))
	for i := range insertedUsrs {
		s.testUsersMap[insertedUsrs[i].Id] = insertedUsrs[i]
		s.testUserIds = append(s.testUserIds, insertedUsrs[i].Id)
	}
}

func (s *UserRepositoryTestSuite) TearDownTest() {
	s.repo.DeleteUsersByCampaign(testOnlyCampaign)
}

func (s *UserRepositoryTestSuite) TestInsertManyUsers() {
	users := []*models.User{
		{Id: uuid.NewString(), Campaigns: []string{testOnlyCampaign}},
		{Id: uuid.NewString(), Campaigns: []string{testOnlyCampaign}},
	}
	userIdsMap := map[string]*models.User{
		users[0].Id: users[0],
		users[1].Id: users[1],
	}
	insertedUsers, err := s.repo.InsertManyUsers(users)

	s.Assert().NoError(err)
	s.Assert().Equal(len(users), len(insertedUsers))
	for i := range insertedUsers {
		s.Assert().NotNil(userIdsMap[insertedUsers[i].Id])
		if s.Assert().NotZero(len(insertedUsers[i].Campaigns)) {
			s.Assert().Equal(testOnlyCampaign, insertedUsers[i].Campaigns[0])
		}
	}
}

func (s *UserRepositoryTestSuite) TestDeleteUsersByIds() {
	err := s.repo.DeleteUsersByIds(s.testUserIds)
	if s.Assert().NoError(err) {
		listResult, err := s.repo.GetListUsersByIds(s.testUserIds)
		s.Assert().NoError(err)
		s.Assert().Equal(0, len(listResult))
	}
}

func (s *UserRepositoryTestSuite) TestDeleteUsersByCampaign() {
	err := s.repo.DeleteUsersByCampaign(testOnlyCampaign)
	if s.Assert().NoError(err) {
		listResult, err := s.repo.GetListUsersByCampaign(testOnlyCampaign)
		s.Assert().NoError(err)
		s.Assert().Equal(0, len(listResult))
	}
}

func (s *UserRepositoryTestSuite) TestGetListUsersByIds() {
	listResult, err := s.repo.GetListUsersByIds(s.testUserIds)
	if s.Assert().NoError(err) && s.Assert().NotEmpty(listResult) {
		for i := range listResult {
			s.Assert().NotNil(s.testUsersMap[listResult[i].Id])
		}
	}
}

func (s *UserRepositoryTestSuite) TestGetListUsersByCampaign() {
	listResult, err := s.repo.GetListUsersByCampaign(testOnlyCampaign)
	if s.Assert().NoError(err) && s.Assert().NotEmpty(listResult) {
		for i := range listResult {
			s.Assert().NotNil(s.testUsersMap[listResult[i].Id])
		}
	}
}

func (s *UserRepositoryTestSuite) TestCountUserDevicesForCampaign() {
	count, err := s.repo.CountUserDevicesForCampaign(testOnlyCampaign)
	if s.Assert().NoError(err) {
		expectedCount := uint64(len(s.testUsersMap) * len(testDeviceTokens))
		s.Assert().Equal(expectedCount, count)
	}
}

func (s *UserRepositoryTestSuite) insertFakeUsers() ([]*models.User, error) {
	n := rand.Intn(5) + 1
	usrs := make([]*models.User, n)
	for i := range n {
		usrs[i] = &models.User{
			Id:              uuid.NewString(),
			Campaigns:       []string{testOnlyCampaign},
			DeviceTokens:    testDeviceTokens,
			EmailVerifiedAt: time.Now().Add(-1 * time.Hour), // 1 hour ago
		}
	}
	return s.repo.InsertManyUsers(usrs)
}
