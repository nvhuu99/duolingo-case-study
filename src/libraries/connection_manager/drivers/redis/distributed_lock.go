package redis

import (
	"context"
	"errors"
	"math/rand"
	"sync/atomic"
	"time"

	events "duolingo/libraries/events/facade"

	"github.com/google/uuid"
	redis_driver "github.com/redis/go-redis/v9"
)

var (
	ErrLockValueEmpty                   = errors.New("the lock value is empty")
	ErrLocksAlreadyAcquired             = errors.New("locks have already acquired")
	ErrLockAcquireTimeout               = errors.New("locks were not acquired before timeout")
	ErrLockReleaseBeforeAcquire         = errors.New("lock released request failure, the locks have not yet been acquired")
	ErrDistributedLockCreateWithoutKeys = errors.New("attempt to create a distributed lock without resources name")
)

type DistributedLock struct {
	client *RedisClient

	lockValue    string
	resourceKeys []string
	acquiredAt   time.Time
	releasedAt   time.Time
	isLocked     atomic.Bool
}

func NewDistributedLock(client *RedisClient, resourceKeys []string) *DistributedLock {
	if len(resourceKeys) == 0 {
		panic(ErrDistributedLockCreateWithoutKeys)
	}
	return &DistributedLock{
		client:       client,
		resourceKeys: resourceKeys,
	}
}

func (lock *DistributedLock) GetLockHeldDuration() time.Duration {
	if lock.releasedAt.After(lock.acquiredAt) {
		return lock.releasedAt.Sub(lock.acquiredAt)
	}
	return 0
}

func (lock *DistributedLock) AcquireLock(ctx context.Context) error {
	if lock.isLocked.Load() {
		return ErrLocksAlreadyAcquired
	}

	// create a new lock value
	newLockValue := uuid.NewString()

	// update lock value if acquired successfully
	defer func() {
		lock.lockValue = ""
		if lock.isLocked.Load() {
			lock.lockValue = newLockValue
		}
	}()

	// try to acquire the locks within timeout
	var acquireErr error

	evt := events.Start(ctx, "conn_manager.redis.acquire_lock", nil)
	defer events.End(evt, true, acquireErr, nil)

	client := lock.client
	timeout := time.After(client.lockAcquireTimeout)
	minWait := client.lockAcquireRetryWaitMin.Milliseconds()
	maxWait := client.lockAcquireRetryWaitMax.Milliseconds()
	for {
		select {
		case <-timeout:
			acquireErr = ErrLockAcquireTimeout
			return acquireErr
		default:
			// acquire lock
			lock.client.ExecuteClosure(ctx, client.lockAcquireTimeout, func(
				timeoutCtx context.Context,
				rdb *redis_driver.Client,
			) error {
				acquireErr = acquireLock(
					timeoutCtx, rdb, newLockValue, lock.resourceKeys, client.lockTTL,
				)
				return acquireErr
			})
			if acquireErr == nil {
				lock.acquiredAt = time.Now()
				lock.isLocked.Store(true)
				return nil
			}
			// failed to acquire the lock, sleep for random wait before retry
			wait := rand.Int63n(maxWait-minWait+1) + minWait
			time.Sleep(time.Duration(wait) * time.Millisecond)
		}
	}
}

func (lock *DistributedLock) ReleaseLock(ctx context.Context) error {
	if !lock.isLocked.Load() {
		return ErrLockReleaseBeforeAcquire
	}
	if lock.lockValue == "" {
		return ErrLockValueEmpty
	}
	var realeaseErr error
	lock.client.ExecuteClosure(ctx, lock.client.lockAcquireTimeout, func(
		timeoutCtx context.Context,
		rdb *redis_driver.Client,
	) error {
		realeaseErr = releaseLock(timeoutCtx, rdb, lock.lockValue, lock.resourceKeys)
		return realeaseErr
	})
	if realeaseErr != nil {
		return realeaseErr
	}
	lock.releasedAt = time.Now()
	lock.isLocked.Store(false)
	return nil
}
