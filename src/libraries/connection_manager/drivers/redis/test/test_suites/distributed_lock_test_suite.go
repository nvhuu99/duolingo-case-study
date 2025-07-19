package test_suites

import (
	"fmt"
	"math/rand"

	redis "duolingo/libraries/connection_manager/drivers/redis"

	"github.com/stretchr/testify/suite"
)

type DistributedLockTestSuite struct {
	suite.Suite
	redisClient *redis.RedisClient
}

func NewDistributedLockTestSuite(redisClient *redis.RedisClient) *DistributedLockTestSuite {
	return &DistributedLockTestSuite{
		redisClient: redisClient,
	}
}

func (s *DistributedLockTestSuite) Test_NewDistributedLock_PanicsOnEmptyKeys() {
	s.Assert().Panics(func() {
		redis.NewDistributedLock(s.redisClient, []string{})
	})
}

func (s *DistributedLockTestSuite) Test_AcquireLock_AcquireSuccessAndLockAreWorking() {
	// Acquiring the test lock
	resourceKeys := s.randomResourceKeys()
	testLock := s.makeLock(resourceKeys)
	acquireErr := testLock.AcquireLock()
	defer testLock.ReleaseLock()
	// Must acquire successfully
	s.Assert().NoError(acquireErr)
	// Must fail before the "testLock" is released for all resource key
	for i := range resourceKeys {
		err := s.makeLock([]string{resourceKeys[i]}).AcquireLock()
		s.Assert().Error(err)
	}
}

func (s *DistributedLockTestSuite) Test_ReleaseLock_AcquireSuccessAfterRelease() {
	resourceKeys := s.randomResourceKeys()
	testLock := s.makeLock(resourceKeys)
	testLock.AcquireLock()
	testLock.ReleaseLock()

	for i := range resourceKeys {
		err := s.makeLock([]string{resourceKeys[i]}).AcquireLock()
		s.Assert().NoError(err)
	}
}

func (s *DistributedLockTestSuite) randomResourceKeys() []string {
	n := rand.Intn(3) + 1
	resourceKeys := make([]string, n)
	for i := range n {
		resourceKeys[i] = fmt.Sprintf("resource_key_%v", i)
	}
	return resourceKeys
}

func (s *DistributedLockTestSuite) makeLock(resourceKeys []string) *redis.DistributedLock {
	return redis.NewDistributedLock(s.redisClient, resourceKeys)
}
