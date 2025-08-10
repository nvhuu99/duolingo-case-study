package redis

import (
	"context"
	connection "duolingo/libraries/connection_manager/drivers/redis"
	distributor "duolingo/libraries/work_distributor"
	"encoding/json"

	"github.com/redis/go-redis/v9"
)

const (
	workloadKeyPrefix              = "work_distributor:workload:"
	assignmentsOfWorkloadKeyPrefix = "work_distributor:workload_assignments:"
)

/*
### Notions:
 1. Some of the operations use transactions without retry on "redis.TxFailedErr",
    due to the ExecuteClosureWithLocks() already ensures atomicity with distributed-lock.
*/
type RedisWorkStorageProxy struct {
	connection.RedisClient
}

func NewRedisWorkStorageProxy(client *connection.RedisClient) *RedisWorkStorageProxy {
	return &RedisWorkStorageProxy{
		RedisClient: *client,
	}
}

func (proxy *RedisWorkStorageProxy) SaveWorkload(
	ctx context.Context, 
	w *distributor.Workload,
) error {
	if validateErr := w.Validate(); validateErr != nil {
		return validateErr
	}
	marshaled, err := json.Marshal(w)
	if err != nil {
		return err
	}
	lockKeys := []string{
		workloadKeyPrefix + w.Id,
	}
	saveErr := proxy.ExecuteClosureWithLocks(ctx, lockKeys, proxy.GetDefaultTimeOut(), func(
		ctx context.Context,
		rdb *redis.Client,
	) error {
		cmd := rdb.Set(ctx, workloadKey(w.Id), string(marshaled), 0)
		_, err := cmd.Result()
		return err
	})

	return saveErr
}

func (proxy *RedisWorkStorageProxy) GetWorkload(
	ctx context.Context,
	workloadId string,
) (*distributor.Workload, error) {
	var marshaled string
	var getErr error
	lockKeys := []string{
		workloadKey(workloadId),
	}
	proxy.ExecuteClosureWithLocks(ctx, lockKeys, proxy.GetDefaultTimeOut(), func(
		ctx context.Context,
		rdb *redis.Client,
	) error {
		marshaled, getErr = rdb.Get(ctx, workloadKey(workloadId)).Result()
		return getErr
	})
	if getErr != nil {
		return nil, errOrRedisNilAlias(getErr, distributor.ErrWorkloadNotExists)
	}

	workload := new(distributor.Workload)
	marshalErr := json.Unmarshal([]byte(marshaled), workload)
	if marshalErr != nil {
		return nil, marshalErr
	}

	return workload, nil
}

func (proxy *RedisWorkStorageProxy) PushAssignmentToQueue(
	ctx context.Context,
	assignment *distributor.Assignment,
) error {
	if validationErr := assignment.Validate(); validationErr != nil {
		return validationErr
	}
	workloadId := assignment.WorkloadId
	lockKeys := []string{
		assignmentsOfWorkloadKey(workloadId),
	}
	saveErr := proxy.ExecuteClosureWithLocks(ctx, lockKeys, proxy.GetDefaultTimeOut(), func(
		ctx context.Context,
		rdb *redis.Client,
	) error {
		assignmentJson, marshalErr := json.Marshal(assignment)
		if marshalErr != nil {
			return marshalErr
		}
		cmd := rdb.RPush(ctx, assignmentsOfWorkloadKey(workloadId), string(assignmentJson))
		_, err := cmd.Result()
		return err
	})
	return saveErr
}

func (proxy *RedisWorkStorageProxy) PopAssignmentFromQueue(
	ctx context.Context,
	workloadId string,
) (
	*distributor.Assignment,
	error,
) {
	assignment := new(distributor.Assignment)
	lockKeys := []string{
		assignmentsOfWorkloadKey(workloadId),
	}
	popErr := proxy.ExecuteClosureWithLocks(ctx, lockKeys, proxy.GetDefaultTimeOut(), func(
		ctx context.Context,
		rdb *redis.Client,
	) error {
		cmd := rdb.LPop(ctx, assignmentsOfWorkloadKey(workloadId))
		assignmentJson, err := cmd.Result()
		if err != nil {
			return err
		}
		unmarshalErr := json.Unmarshal([]byte(assignmentJson), assignment)
		return unmarshalErr
	})
	if popErr != nil {
		// return no error instead of RedisNil (indicates the queue is empty)
		return nil, errOrRedisNilAlias(popErr, nil)
	}
	// got an assignment
	return assignment, nil
}

func (proxy *RedisWorkStorageProxy) GetAndUpdateWorkload(
	ctx context.Context, 
	workloadId string,
	modifier func(*distributor.Workload) error,
) error {
	lockKeys := []string{
		workloadKey(workloadId),
	}
	return proxy.ExecuteClosureWithLocks(ctx, lockKeys, proxy.GetDefaultTimeOut(), func(
		ctx context.Context,
		rdb *redis.Client,
	) error {
		// ensure atomicity when get/set workload
		return rdb.Watch(ctx, func(tx *redis.Tx) error {
			// get workload from storage
			workloadJson, readErr := tx.Get(ctx, workloadKey(workloadId)).Result()
			if readErr != nil {
				return errOrRedisNilAlias(readErr, distributor.ErrWorkloadNotExists)
			}
			// call the modifier
			workload := unmarshalWorkloadIgnoreErr(workloadJson)
			if modifyErr := modifier(workload); modifyErr != nil {
				return modifyErr
			}
			updated := marshalWorkloadIgnoreErr(workload)
			// then store it back
			_, pipelineErr := tx.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
				pipe.Set(ctx, workloadKey(workloadId), updated, 0)
				return nil
			})
			return pipelineErr
		})
	})
}

func (proxy *RedisWorkStorageProxy) DeleteWorkloadAndAssignments(
	ctx context.Context,
	workloadId string,
) error {
	lockKeys := []string{
		workloadKey(workloadId),
		assignmentsOfWorkloadKey(workloadId),
	}
	return proxy.ExecuteClosureWithLocks(ctx, lockKeys, proxy.GetDefaultTimeOut(), func(
		ctx context.Context,
		rdb *redis.Client,
	) error {
		// ensure atomicity when delete workload & assignments
		return rdb.Watch(ctx, func(tx *redis.Tx) error {
			// check workload existence
			_, readErr := tx.Get(ctx, workloadKey(workloadId)).Result()
			if readErr != nil {
				return errOrRedisNilAlias(readErr, distributor.ErrWorkloadNotExists)
			}
			// delete workload & assignments
			_, pipelineErr := tx.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
				pipe.Del(ctx, workloadKey(workloadId))
				pipe.Del(ctx, assignmentsOfWorkloadKey(workloadId))
				return nil
			})
			return pipelineErr
		})
	})
}
