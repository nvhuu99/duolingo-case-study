package suite

import (
	"duolingo/repositories/campaign_message/user_repository"
	"duolingo/repositories/campaign_message/user_repository/models"
	"math/rand"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
)

const (
	TEST_ONLY_CAMPAIGN = "TEST_ONLY_CAMPAIGN"
)

type UserRepositorySuite struct {
	suite.Suite
	repo user_repository.UserRepository
}

func (s *UserRepositorySuite) TearDownTest() {
	if usrs, err := s.repo.GetListUsersByCampaign(TEST_ONLY_CAMPAIGN); err != nil {
		s.repo.DeleteUsers(usrs)
	}
}

func (s *UserRepositorySuite) TestInsertManyUsers() {
	usrs := []*models.User{
		{Id: uuid.NewString(), Campaigns: []string{TEST_ONLY_CAMPAIGN}},
		{Id: uuid.NewString(), Campaigns: []string{TEST_ONLY_CAMPAIGN}},
	}
	insertedUsrs, err := s.repo.InsertManyUsers(usrs)
	s.Assert().NoError(err)
	s.Assert().Equal(len(usrs), len(insertedUsrs))
	for i := range insertedUsrs {
		if s.Assert().NotZero(len(insertedUsrs[i].Campaigns)) {
			s.Assert().Equal(TEST_ONLY_CAMPAIGN, insertedUsrs[i].Campaigns[0])
		}
		s.Assert().Equal(usrs[i].Id, insertedUsrs[i].Id)
	}
}

func (s *UserRepositorySuite) TestDeleteUsers() {
	insertedUsrs, err := s.insertFakeUsers()
	if s.Assert().NoError(err) {
		usrIds := make([]string, len(insertedUsrs))
		for i := range insertedUsrs {
			usrIds[i] = insertedUsrs[i].Id
		}
		err = s.repo.DeleteUsers(insertedUsrs)
		if s.Assert().NoError(err) {
			listResult, err := s.repo.GetListUsersByIds(usrIds)
			s.Assert().NoError(err)
			s.Assert().Equal(0, len(listResult))
		}
	}
}

func (s *UserRepositorySuite) TestDeleteByUserIds() {
	insertedUsrs, err := s.insertFakeUsers()
	if s.Assert().NoError(err) {
		usrIds := make([]string, len(insertedUsrs))
		for i := range insertedUsrs {
			usrIds[i] = insertedUsrs[i].Id
		}
		err = s.repo.DeleteUsersByIds(usrIds)
		if s.Assert().NoError(err) {
			listResult, err := s.repo.GetListUsersByIds(usrIds)
			s.Assert().NoError(err)
			s.Assert().Equal(0, len(listResult))
		}
	}
}

func (s *UserRepositorySuite) TestGetListUsersByIds() {
	insertedUsrs, err := s.insertFakeUsers()
	if s.Assert().NoError(err) {
		usrIds := make([]string, len(insertedUsrs))
		for i := range insertedUsrs {
			usrIds[i] = insertedUsrs[i].Id
		}
		listResult, err := s.repo.GetListUsersByIds(usrIds)
		if s.Assert().NoError(err) {
			for i := range insertedUsrs {
				s.Assert().Equal(insertedUsrs[i].Id, listResult[i].Id)
			}
		}
	}
}

func (s *UserRepositorySuite) TestGetListUsersByCampaign() {
	insertedUsrs, err := s.insertFakeUsers()
	if s.Assert().NoError(err) {
		listResult, err := s.repo.GetListUsersByCampaign(TEST_ONLY_CAMPAIGN)
		if s.Assert().NoError(err) {
			for i := range insertedUsrs {
				s.Assert().Equal(insertedUsrs[i].Id, listResult[i].Id)
			}
		}
	}
}

func (s *UserRepositorySuite) TestCountUserDevicesForCampaign() {
	insertedUsrs, err := s.insertFakeUsers()
	if s.Assert().NoError(err) {
		count, err := s.repo.CountUserDevicesForCampaign(TEST_ONLY_CAMPAIGN)
		if s.Assert().NoError(err) {
			s.Assert().Equal(len(insertedUsrs)*2, count)
		}
	}
}

func (s *UserRepositorySuite) insertFakeUsers() ([]*models.User, error) {
	n := rand.Intn(5) + 1
	usrs := make([]*models.User, n)
	for i := range n {
		usrs[i] = &models.User{
			Id:           uuid.NewString(),
			Campaigns:    []string{TEST_ONLY_CAMPAIGN},
			DeviceTokens: []string{"fake_token_1", "fake_token_2"},
		}
	}
	return s.repo.InsertManyUsers(usrs)
}
