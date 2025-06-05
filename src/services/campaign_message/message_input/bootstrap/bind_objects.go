package bootstrap

import (
	"duolingo/libraries/pub_sub"
	"duolingo/libraries/work_distributor"
	"duolingo/repositories/campaign_message/user_repository"
)

func GetPublisher() pub_sub.Publisher {
	// TODO: need implementation
	return nil
}

func GetWorkDistributor() work_distributor.WorkDistributor {
	// TODO: need implementation
	return nil
}

func GetUserRepository() user_repository.UserRepository {
	// TODO: need implementation
	return nil
}
