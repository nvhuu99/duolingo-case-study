package test_suites

import (
	"sort"
	"time"

	container "duolingo/libraries/dependencies_container"
	"duolingo/models"
	user_repo "duolingo/repositories/user_repository/external"
	"duolingo/repositories/user_repository/external/commands"
	"duolingo/test/fixtures/data"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
)

type UserRepositoryTestSuite struct {
	suite.Suite
	factory user_repo.UserRepoFactory
	repo    user_repo.UserRepository
}

func NewUserRepositoryTestSuite() *UserRepositoryTestSuite {
	factory := container.MustResolve[user_repo.UserRepoFactory]()
	repo := container.MustResolve[user_repo.UserRepository]()
	return &UserRepositoryTestSuite{
		factory: factory,
		repo:    repo,
	}
}

func (s *UserRepositoryTestSuite) SetupTest() {
	s.repo.InsertManyUsers(data.TestUsers)
	s.repo.InsertManyUsers(data.TestUsersEmailUnverified)
}

func (s *UserRepositoryTestSuite) TearDownTest() {
	s.repo.DeleteUsersByIds(data.TestUserIds)
	s.repo.DeleteUsersByIds(data.TestUsersEmailUnverifiedIds)
}

func (s *UserRepositoryTestSuite) Test_InsertManyUsers() {
	users := make([]*models.User, 2)
	users[0] = &models.User{
		Id:        uuid.NewString(),
		Lastname:  "usr_1_lastname",
		Firstname: "usr_1_firstname",
		Username:  "usr_1_username",
		Email:     "usr_1@demo.com",
		Campaigns: []string{
			"usr_1_campaign_1",
			"usr_1_campaign_2",
		},
		Devices: []*models.UserDevice{
			{Platform: "android", Token: "usr_1_android_token"},
			{Platform: "ios", Token: "usr_1_ios_token"},
		},
		NativeLanguage:  models.LanguageJP,
		Membership:      models.MembershipFreeTier,
		EmailVerifiedAt: time.Now().UTC().Add(-1 * time.Hour),
	}
	users[1] = &models.User{
		Id:        uuid.NewString(),
		Lastname:  "usr_2_lastname",
		Firstname: "usr_2_firstname",
		Username:  "usr_2_username",
		Email:     "usr_2@demo.com",
		Campaigns: []string{
			"usr_2_campaign_1",
			"usr_2_campaign_2",
		},
		Devices: []*models.UserDevice{
			{Platform: "android", Token: "usr_2_android_token"},
			{Platform: "ios", Token: "usr_2_ios_token"},
		},
		NativeLanguage:  models.LanguageVN,
		Membership:      models.MembershipSubscription,
		EmailVerifiedAt: time.Now().UTC().Add(-3 * time.Hour),
	}
	sort.Slice(users, func(i, j int) bool {
		return users[i].Id < users[j].Id
	})

	insertedUsers, insertErr := s.repo.InsertManyUsers(users)
	defer s.repo.DeleteUsersByIds([]string{
		insertedUsers[0].Id,
		insertedUsers[1].Id,
	})

	if !s.Assert().NoError(insertErr) {
		return
	}

	query := s.factory.MakeListUsersCommand()
	query.SetFilterIds([]string{users[0].Id, users[1].Id})
	query.SetSortById(commands.OrderASC)
	listResults, _ := s.repo.GetListUsers(query)

	s.Assert().True(users[0].Equal(listResults[0]))
	s.Assert().True(users[1].Equal(listResults[1]))
}

func (s *UserRepositoryTestSuite) Test_DeleteUsers_ByIds() {
	err := s.repo.DeleteUsersByIds(data.TestUserIds)

	// confirm no error
	s.Assert().NoError(err)

	// confirm actually deleted
	listResult1, _ := s.repo.GetListUsersByIds(data.TestUserIds)
	s.Assert().Equal(0, len(listResult1))

	// confirm deleted correctly
	listResult2, _ := s.repo.GetListUsersByIds(data.TestUsersEmailUnverifiedIds)
	s.Assert().Equal(len(data.TestUsersEmailUnverifiedIds), len(listResult2))
}

func (s *UserRepositoryTestSuite) Test_GetListUsers_ByIds() {
	listResult1, err1 := s.repo.GetListUsersByIds(data.TestUserIds)
	if s.Assert().NoError(err1) && s.Assert().NotEmpty(listResult1) {
		for i := range data.TestUsers {
			s.Assert().True(data.TestUsers[i].Equal(listResult1[i]))
		}
	}
}

func (s *UserRepositoryTestSuite) Test_GetListUsers_ByCampaign() {
	query1 := s.factory.MakeListUsersCommand()
	query1.SetFilterCampaign(data.TestCampaignPrimary)
	query1.SetSortById(commands.OrderASC)
	listResult1, err1 := s.repo.GetListUsers(query1)

	if s.Assert().NoError(err1) && s.Assert().NotEmpty(listResult1) {
		campaignUserIds := data.TestCampaignUserIdsMap[data.TestCampaignPrimary]
		for i := range campaignUserIds {
			s.Assert().Equal(listResult1[i].Id, campaignUserIds[i])
		}
	}

	query2 := s.factory.MakeListUsersCommand()
	query2.SetFilterCampaign(data.TestCampaignSecondary)
	query2.SetSortById(commands.OrderASC)
	listResult2, err2 := s.repo.GetListUsers(query2)

	if s.Assert().NoError(err2) && s.Assert().NotEmpty(listResult2) {
		campaignUserIds := data.TestCampaignUserIdsMap[data.TestCampaignSecondary]
		for i := range campaignUserIds {
			s.Assert().Equal(listResult2[i].Id, campaignUserIds[i])
		}
	}
}

func (s *UserRepositoryTestSuite) Test_GetListUsers_EmailVerifiedOnly() {
	query := s.factory.MakeListUsersCommand()
	query.SetFilterOnlyEmailVerified()
	query.SetSortById(commands.OrderASC)
	listResult, err := s.repo.GetListUsers(query)

	if s.Assert().NoError(err) && s.Assert().NotEmpty(listResult) {
		for i := range data.TestUserIds {
			s.Assert().Equal(listResult[i].Id, data.TestUserIds[i])
		}
	}
}

func (s *UserRepositoryTestSuite) Test_GetListUsers_By_Campaign_And_EmailVerified() {
	query := s.factory.MakeListUsersCommand()
	query.SetFilterCampaign(data.TestCampaignSecondary)
	query.SetFilterOnlyEmailVerified()
	query.SetSortById(commands.OrderASC)

	listResult, err := s.repo.GetListUsers(query)
	if s.Assert().NoError(err) && s.Assert().NotEmpty(listResult) {
		usrIds := data.TestCampaignUserIdsMap[data.TestCampaignSecondary]
		s.Assert().Equal(len(listResult), len(usrIds))
		for i := range usrIds {
			s.Assert().Equal(usrIds[i], listResult[i].Id)
		}
	}
}

func (s *UserRepositoryTestSuite) Test_GetListUsers_WithPagination() {
	size := 2
	total := (len(data.TestUsers) + 1) / size // ceiling(len/size)
	for page := range total {
		query := s.factory.MakeListUsersCommand()
		query.SetFilterIds(data.TestUserIds)
		query.SetSortById(commands.OrderASC)
		query.SetPagination(int64(page*size), int64(size))

		listResult, err := s.repo.GetListUsers(query)
		if s.Assert().NoError(err) && s.Assert().NotEmpty(listResult) {
			for i := range size {
				if j := page*size + i; j < len(data.TestUserIds) {
					s.Assert().Equal(data.TestUserIds[j], listResult[i].Id)
				}
			}
		}
	}
}

func (s *UserRepositoryTestSuite) Test_GetListUsers_WithSortById() {
	query1 := s.factory.MakeListUsersCommand()
	query1.SetFilterIds(data.TestUserIds)
	query1.SetSortById(commands.OrderASC)
	listResult1, err1 := s.repo.GetListUsers(query1)
	if s.Assert().NoError(err1) && s.Assert().NotEmpty(listResult1) {
		for i := range data.TestUsers {
			s.Assert().Equal(data.TestUserIds[i], listResult1[i].Id)
		}
	}

	sort.Slice(data.TestUserIds, func(i, j int) bool {
		return data.TestUserIds[i] > data.TestUserIds[j]
	})
	query2 := s.factory.MakeListUsersCommand()
	query2.SetFilterIds(data.TestUserIds)
	query2.SetSortById(commands.OrderDESC)
	listResult2, err2 := s.repo.GetListUsers(query2)
	if s.Assert().NoError(err2) && s.Assert().NotEmpty(listResult2) {
		for i := range data.TestUsers {
			s.Assert().Equal(data.TestUserIds[i], listResult2[i].Id)
		}
	}
}

func (s *UserRepositoryTestSuite) Test_GetListUserDevices() {
	size := 2
	total := (len(data.TestDevices) + 1) / size // ceiling(len/size)
	for page := range total {
		query := s.factory.MakeListUserDevicesCommand()
		query.SetFilterCampaign(data.TestCampaignPrimary)
		query.SetSortById(commands.OrderASC)
		query.SetPagination(int64(page*size), int64(size))
		devices, err := s.repo.GetListUserDevices(query)

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

func (s *UserRepositoryTestSuite) Test_AggregateUsers_SumUserDevices_Of_CampaignMessageReceivers() {
	aggregation := s.factory.MakeAggregateUsersCommand()
	aggregation.SetFilterCampaign(data.TestCampaignPrimary)
	aggregation.SetFilterOnlyEmailVerified()
	aggregation.AddAggregationSumUserDevices()

	result, err := s.repo.AggregateUsers(aggregation)
	if s.Assert().NotNil(result) && s.Assert().NoError(err) {
		expectedCount := int64(len(data.TestDevices))
		actualCount := result.GetCountUserDevices()
		s.Assert().Equal(expectedCount, actualCount)
	}
}
