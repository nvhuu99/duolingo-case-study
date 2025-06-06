package suite

import (
	user_repository "duolingo/repositories/campaign_message/user_repository/external"
	"duolingo/repositories/campaign_message/user_repository/models"
	"math/rand"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
)

var (
	testOnlyCampaign = "testOnlyCampaign"
	testDeviceTokens = []string{"fake_token_1", "fake_token_2"}
)

type UserRepositorySuite struct {
	suite.Suite
	repo         user_repository.UserRepository
	testUsersMap map[string]*models.User
	testUserIds  []string
}

func NewUserRepositorySuite(repo user_repository.UserRepository) *UserRepositorySuite {
	return &UserRepositorySuite{
		repo: repo,
	}
}

func (s *UserRepositorySuite) SetupTest() {
	insertedUsrs, err := s.insertFakeUsers()
	if err == nil {
		s.testUsersMap = make(map[string]*models.User)
		s.testUserIds = make([]string, len(insertedUsrs))
		for i := range insertedUsrs {
			s.testUsersMap[insertedUsrs[i].Id] = insertedUsrs[i]
			s.testUserIds = append(s.testUserIds, insertedUsrs[i].Id)
		}
	}
}

func (s *UserRepositorySuite) TearDownTest() {
	s.repo.DeleteUsersByCampaign(testOnlyCampaign)
}

func (s *UserRepositorySuite) TestInsertManyUsers() {
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

func (s *UserRepositorySuite) TestDeleteUsersByIds() {
	err := s.repo.DeleteUsersByIds(s.testUserIds)
	if s.Assert().NoError(err) {
		listResult, err := s.repo.GetListUsersByIds(s.testUserIds)
		s.Assert().NoError(err)
		s.Assert().Equal(0, len(listResult))
	}
}

func (s *UserRepositorySuite) TestDeleteUsersByCampaign() {
	err := s.repo.DeleteUsersByCampaign(testOnlyCampaign)
	if s.Assert().NoError(err) {
		listResult, err := s.repo.GetListUsersByCampaign(testOnlyCampaign)
		s.Assert().NoError(err)
		s.Assert().Equal(0, len(listResult))
	}
}

func (s *UserRepositorySuite) TestGetListUsersByIds() {
	listResult, err := s.repo.GetListUsersByIds(s.testUserIds)
	if s.Assert().NoError(err) && s.Assert().NotEmpty(listResult) {
		for i := range listResult {
			s.Assert().NotNil(s.testUsersMap[listResult[i].Id])
		}
	}
}

func (s *UserRepositorySuite) TestGetListUsersByCampaign() {
	listResult, err := s.repo.GetListUsersByCampaign(testOnlyCampaign)
	if s.Assert().NoError(err) && s.Assert().NotEmpty(listResult) {
		for i := range listResult {
			s.Assert().NotNil(s.testUsersMap[listResult[i].Id])
		}
	}
}

func (s *UserRepositorySuite) TestCountUserDevicesForCampaign() {
	count, err := s.repo.CountUserDevicesForCampaign(testOnlyCampaign)
	if s.Assert().NoError(err) {
		s.Assert().Equal(len(s.testUsersMap)*len(testDeviceTokens), count)
	}
}

func (s *UserRepositorySuite) insertFakeUsers() ([]*models.User, error) {
	n := rand.Intn(5) + 1
	usrs := make([]*models.User, n)
	for i := range n {
		usrs[i] = &models.User{
			Id:           uuid.NewString(),
			Campaigns:    []string{testOnlyCampaign},
			DeviceTokens: testDeviceTokens,
		}
	}
	return s.repo.InsertManyUsers(usrs)
}
