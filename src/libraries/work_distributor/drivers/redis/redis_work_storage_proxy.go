package redis

import (
	"context"
	connection "duolingo/libraries/connection_manager/drivers/redis"
	events "duolingo/libraries/events/facade"
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
	saveEvt := events.Start(ctx, "work_dist.proxy.redis.save_workload", nil)

	if validateErr := w.Validate(); validateErr != nil {
		events.Failed(saveEvt, validateErr, nil)
		return validateErr
	}

	marshaled, err := json.Marshal(w)
	if err != nil {
		events.Failed(saveEvt, err, nil)
		return err
	}

	lockKeys := []string{
		workloadKeyPrefix + w.Id,
	}
	saveErr := proxy.ExecuteClosureWithLocks(saveEvt.Context(), lockKeys, proxy.GetDefaultTimeOut(), func(
		timeoutCtx context.Context,
		rdb *redis.Client,
	) error {
		cmd := rdb.Set(timeoutCtx, workloadKey(w.Id), string(marshaled), 0)
		_, err := cmd.Result()
		return err
	})

	events.End(saveEvt, true, saveErr, nil)

	return saveErr
}

func (proxy *RedisWorkStorageProxy) GetWorkload(
	ctx context.Context,
	workloadId string,
) (*distributor.Workload, error) {
	var marshaled string
	var err error

	getEvt := events.Start(ctx, "work_dist.proxy.redis.get_workload", nil)
	defer events.End(getEvt, true, err, nil)

	lockKeys := []string{
		workloadKey(workloadId),
	}
	err = proxy.ExecuteClosureWithLocks(getEvt.Context(), lockKeys, proxy.GetDefaultTimeOut(), func(
		timeoutCtx context.Context,
		rdb *redis.Client,
	) error {
		var cmdErr error
		cmd := rdb.Get(timeoutCtx, workloadKey(workloadId))
		marshaled, cmdErr = cmd.Result()
		return cmdErr
	})
	if err != nil {
		err = errOrRedisNilAlias(err, distributor.ErrWorkloadNotExists)
		return nil, err
	}

	workload := new(distributor.Workload)
	err = json.Unmarshal([]byte(marshaled), workload)
	if err != nil {
		return nil, err
	}

	return workload, nil
}

func (proxy *RedisWorkStorageProxy) PushAssignmentToQueue(
	ctx context.Context,
	assignment *distributor.Assignment,
) error {
	evt := events.Start(ctx, "work_dist.proxy.push_assignment", nil)

	if validationErr := assignment.Validate(); validationErr != nil {
		events.Failed(evt, validationErr, nil)
		return validationErr
	}

	workloadId := assignment.WorkloadId
	lockKeys := []string{
		assignmentsOfWorkloadKey(workloadId),
	}
	saveErr := proxy.ExecuteClosureWithLocks(evt.Context(), lockKeys, proxy.GetDefaultTimeOut(), func(
		timeoutCtx context.Context,
		rdb *redis.Client,
	) error {
		assignmentJson, marshalErr := json.Marshal(assignment)
		if marshalErr != nil {
			return marshalErr
		}
		cmd := rdb.RPush(timeoutCtx, lockKeys[0], string(assignmentJson))
		_, err := cmd.Result()
		return err
	})

	events.End(evt, true, saveErr, nil)

	return saveErr
}

func (proxy *RedisWorkStorageProxy) PopAssignmentFromQueue(
	ctx context.Context,
	workloadId string,
) (*distributor.Assignment, error) {
	var err error

	evt := events.Start(ctx, "work_dist.proxy.redis.pop_assignment", nil)
	defer events.End(evt, true, err, nil)

	assignment := new(distributor.Assignment)
	lockKeys := []string{
		assignmentsOfWorkloadKey(workloadId),
	}
	err = proxy.ExecuteClosureWithLocks(evt.Context(), lockKeys, proxy.GetDefaultTimeOut(), func(
		timeoutCtx context.Context,
		rdb *redis.Client,
	) error {
		cmd := rdb.LPop(timeoutCtx, lockKeys[0])
		assignmentJson, err := cmd.Result()
		if err != nil {
			return err
		}
		unmarshalErr := json.Unmarshal([]byte(assignmentJson), assignment)
		return unmarshalErr
	})
	if err != nil {
		// return no error instead of RedisNil (indicates the queue is empty)
		err = errOrRedisNilAlias(err, nil)
		return nil, err
	}
	// got an assignment
	return assignment, nil
}

func (proxy *RedisWorkStorageProxy) GetAndUpdateWorkload(
	ctx context.Context,
	workloadId string,
	modifier func(*distributor.Workload) error,
) error {
	var err error

	evt := events.Start(ctx, "work_dist.proxy.redis.get_and_update_workload", nil)
	defer events.End(evt, true, err, nil)

	lockKeys := []string{
		workloadKey(workloadId),
	}
	err = proxy.ExecuteClosureWithLocks(evt.Context(), lockKeys, proxy.GetDefaultTimeOut(), func(
		timeoutCtx context.Context,
		rdb *redis.Client,
	) error {
		// ensure atomicity when get/set workload
		return rdb.Watch(timeoutCtx, func(tx *redis.Tx) error {
			// get workload from storage
			workloadJson, readErr := tx.Get(timeoutCtx, lockKeys[0]).Result()
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
			_, pipelineErr := tx.TxPipelined(timeoutCtx, func(pipe redis.Pipeliner) error {
				pipe.Set(timeoutCtx, workloadKey(workloadId), updated, 0)
				return nil
			})
			return pipelineErr
		})
	})

	return err
}

func (proxy *RedisWorkStorageProxy) DeleteWorkloadAndAssignments(
	ctx context.Context,
	workloadId string,
) error {
	var err error

	evt := events.Start(ctx, "work_dist.proxy.redis.delete_workload", nil)
	defer events.End(evt, true, err, nil)

	lockKeys := []string{
		workloadKey(workloadId),
		assignmentsOfWorkloadKey(workloadId),
	}
	err = proxy.ExecuteClosureWithLocks(evt.Context(), lockKeys, proxy.GetDefaultTimeOut(), func(
		timeoutCtx context.Context,
		rdb *redis.Client,
	) error {
		// ensure atomicity when delete workload & assignments
		return rdb.Watch(timeoutCtx, func(tx *redis.Tx) error {
			// check workload existence
			_, readErr := tx.Get(timeoutCtx, workloadKey(workloadId)).Result()
			if readErr != nil {
				return errOrRedisNilAlias(readErr, distributor.ErrWorkloadNotExists)
			}
			// delete workload & assignments
			_, pipelineErr := tx.TxPipelined(timeoutCtx, func(pipe redis.Pipeliner) error {
				pipe.Del(timeoutCtx, workloadKey(workloadId))
				pipe.Del(timeoutCtx, assignmentsOfWorkloadKey(workloadId))
				return nil
			})
			return pipelineErr
		})
	})

	return err
}
