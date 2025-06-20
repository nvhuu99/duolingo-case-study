package test_suites

import (
	"duolingo/models"
	user_repo "duolingo/repositories/user_repository/external"
	"math/rand"
	"time"

	"github.com/stretchr/testify/suite"
)

var (
	testOnlyCampaign = "testOnlyCampaign"
	testDevices      = []*models.UserDevice{
		{Token: "fake_token_1"},
		{Token: "fake_token_2"},
	}
)

type UserRepositoryTestSuite struct {
	suite.Suite
	factory      user_repo.UserRepoFactory
	repo         user_repo.UserRepository
	testUsersMap map[string]*models.User
	testUserIds  []string
}

func NewUserRepositoryTestSuite(factory user_repo.UserRepoFactory) *UserRepositoryTestSuite {
	return &UserRepositoryTestSuite{
		factory: factory,
		repo:    factory.MakeUserRepo(),
	}
}

func (s *UserRepositoryTestSuite) SetupTest() {
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

func (s *UserRepositoryTestSuite) TearDownTest() {
	s.repo.DeleteUsersByIds(s.testUserIds)
}

func (s *UserRepositoryTestSuite) Test_InsertManyUsers() {
	users := []*models.User{
		{Campaigns: []string{testOnlyCampaign}},
		{Campaigns: []string{testOnlyCampaign}},
	}
	insertedUsers, err := s.repo.InsertManyUsers(users)
	defer s.repo.DeleteUsersByIds([]string{
		insertedUsers[0].Id,
		insertedUsers[1].Id,
	})

	s.Assert().NoError(err)
	s.Assert().Equal(len(users), len(insertedUsers))
	for i := range insertedUsers {
		if s.Assert().NotZero(len(insertedUsers[i].Campaigns)) {
			s.Assert().Equal(testOnlyCampaign, insertedUsers[i].Campaigns[0])
		}
	}
}

func (s *UserRepositoryTestSuite) Test_DeleteUsersByIds() {
	err := s.repo.DeleteUsersByIds(s.testUserIds)
	if s.Assert().NoError(err) {
		listResult, err := s.repo.GetListUsersByIds(s.testUserIds)
		s.Assert().NoError(err)
		s.Assert().Equal(0, len(listResult))
	}
}

func (s *UserRepositoryTestSuite) Test_GetListUsersByIds() {
	listResult, err := s.repo.GetListUsersByIds(s.testUserIds)
	if s.Assert().NoError(err) && s.Assert().NotEmpty(listResult) {
		for i := range listResult {
			s.Assert().NotNil(s.testUsersMap[listResult[i].Id])
		}
	}
}

func (s *UserRepositoryTestSuite) Test_GetListUsers_By_Campaign_And_Email_Verified() {
	command := s.factory.MakeListUsersCommand()
	command.SetFilterCampaign(testOnlyCampaign)
	command.SetFilterOnlyEmailVerified()

	listResult, err := s.repo.GetListUsers(command)
	if s.Assert().NoError(err) && s.Assert().NotEmpty(listResult) {
		s.Assert().Equal(len(listResult), len(s.testUserIds))
		for i := range listResult {
			s.Assert().NotNil(s.testUsersMap[listResult[i].Id])
		}
	}
}

func (s *UserRepositoryTestSuite) Test_AggregateUsers_SumUserDevices_Of_CampaignMessageReceivers() {
	aggregation := s.factory.MakeAggregateUsersCommand()
	aggregation.SetFilterCampaign(testOnlyCampaign)
	aggregation.SetFilterOnlyEmailVerified()
	aggregation.AddAggregationSumUserDevices()

	result, err := s.repo.AggregateUsers(aggregation)
	if s.Assert().NotNil(result) && s.Assert().NoError(err) {
		expectedCount := uint64(len(s.testUsersMap) * len(testDevices))
		actualCount := result.GetCountUserDevices()

		s.Assert().Equal(expectedCount, actualCount)
	}
}

func (s *UserRepositoryTestSuite) insertFakeUsers() ([]*models.User, error) {
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
