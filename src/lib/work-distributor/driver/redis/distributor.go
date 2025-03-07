package redisdistributor

import (
	"context"
	"encoding/json"
	"fmt"
	// "log"
	"strconv"
	"time"

	wd "duolingo/lib/work-distributor"

	redis "github.com/redis/go-redis/v9"
)

type RedisDistributor struct {
	rdb		*redis.Client
	opts	*wd.DistributorOptions

	name		string
	workload	*wd.Workload

	ctx     context.Context
}

func NewRedisDistributor(ctx context.Context, name string, opts *wd.DistributorOptions) (*RedisDistributor, error) {
	d := RedisDistributor{}
	d.name = name
	d.ctx = ctx
	if opts == nil {
		opts = wd.DefaultDistributorOptions()
	}
	d.opts = opts

	return &d, nil
}

func (d *RedisDistributor) SetConnection(host string, port string) error {
	opt, err := redis.ParseURL(fmt.Sprintf("redis://%v:%v", host, port))
	if err != nil {
		return wd.NewError(wd.ConnectionFailure,err, "", "", "" )
	}
	d.rdb = redis.NewClient(opt)

	return nil
}

func (d *RedisDistributor) PurgeData() error {
	lockVal, err := d.acquireLock("workloads")
	if err != nil {
		return wd.NewError(wd.LockErr, err, d.name, "", "")
	}
	defer d.releaseLock(lockVal, "workloads")
	// Get the workload list of this distributor
	workloads, err := d.rdb.HKeys(d.ctx, d.key("workloads")).Result()
	if err != nil && err != redis.Nil {
		return wd.NewError(wd.PurgeFailure, err, d.name, "", "")
	}
	// Loop over workloads and purge all data (including the locks)
	for _, name := range workloads {
		if err := d.SwitchToWorkload(name); err != nil {
			return wd.NewError(wd.PurgeFailure, err, d.name, name, "")
		}
		lockVal, err := d.acquireLock("assignments", "assignment_index", "available")
		if err != nil {
			return wd.NewError(wd.LockErr, err, d.name, "", "")
		}
		err = d.rdb.Del(d.ctx, d.key("assignment_index"), d.key("available"), d.key("assignments")).Err()
		if err != nil {
			return wd.NewError(wd.PurgeFailure, err, d.name, "", "")
		}
		d.releaseLock(lockVal, "assignments", "assignment_index", "available")
	}
	// Delete all workloads
	err = d.rdb.Del(d.ctx, d.key("workloads")).Err()
	if err != nil {
		return wd.NewError(wd.PurgeFailure, err, d.name, "", "")
	}

	return nil
}

func (d *RedisDistributor) WorkloadExists(workloadName string) (bool, error) {
	lockVal, err := d.acquireLock("workloads")
	if err != nil {
		return false, wd.NewError(wd.LockErr, err, d.name, "", "")
	}
	defer d.releaseLock(lockVal, "workloads")

	exist, err := d.rdb.HExists(d.ctx, d.key("workloads"), workloadName).Result()
	if err != nil {
		return false, wd.NewError(wd.UnKnownErr, err, d.name, workloadName, "")
	}
	if exist {
		return true, nil
	}
	return false, nil
}

func (d *RedisDistributor) RegisterWorkLoad(workload *wd.Workload) error {
	if ! workload.ValidAttributes() {
		return wd.NewError(wd.WorkloadAttributesInvalid, nil, d.name, "", "")
	}
	lockVal, err := d.acquireLock("workloads")
	if err != nil {
		return wd.NewError(wd.LockErr, err, d.name, "", "")
	}
	defer d.releaseLock(lockVal, "workloads")

	// Skip if the workload has already registered
	exist, err := d.WorkloadExists(workload.Name)
	if err != nil {
		return wd.NewError(wd.WorkloadCreationErr, err, d.name, workload.Name, "")
	}
	if exist {
		return nil
	}
	// Register the workload
	str, _ := json.Marshal(workload)
	err = d.rdb.HSet(d.ctx, d.key("workloads"), workload.Name, string(str)).Err()
	if err != nil {
		return wd.NewError(wd.WorkloadCreationErr, err, d.name, workload.Name, "")
	}

	return nil
}

func (d *RedisDistributor) SwitchToWorkload(name string) error {
	lockVal, err := d.acquireLock("workloads")
	if err != nil {
		return wd.NewError(wd.LockErr, err, d.name, "", "")
	}
	defer d.releaseLock(lockVal, "workloads")

	result, err := d.rdb.HGet(d.ctx, d.key("workloads"), name).Result()
	if err != nil && err != redis.Nil {
		return wd.NewError(wd.WorkloadSwitchErr, err, d.name, name, "")
	}
	if err == redis.Nil {
		return wd.NewError(wd.WorkloadNotFound, nil, d.name, name, "")
	}

	json.Unmarshal([]byte(result), &d.workload)

	return nil
}

func (d *RedisDistributor) Next() (*wd.Assignment, error) {
	if d.workload == nil {
		return nil, wd.NewError(wd.WorkloadNotSet, nil, d.name, "", "")
	}
	// Get next distribution start & end
	next, err := d.available()
	if err != nil {
		return nil, err
	}
	// Create assignment
	assignment, err := d.assign(next[0], next[1])
	if err != nil {
		return nil, err
	}
	// Should resume unfinished assignment
	if assignment.Progress >= assignment.Start {
		assignment.Start = assignment.Progress + 1
	}
	// Got an uncommit assignment, commit and retry
	if assignment.Start >= assignment.End {
		d.Commit(assignment.Id)
		return d.Next()
	}

	return assignment, nil
}

func (d *RedisDistributor) Progress(assignmentId string, newVal int) error {
	lockVal, err := d.acquireLock("assignments")
	if err != nil {
		return wd.NewError(wd.LockErr, err, d.name, "", "")
	}
	defer d.releaseLock(lockVal, "assignments")

	var assignment wd.Assignment
	result, err := d.rdb.HGet(d.ctx, d.key("assignments"), assignmentId).Result()
	if err != nil && err != redis.Nil {
		return wd.NewError(wd.WorkloadAssignmentErr, err, d.name, d.workload.Name, assignmentId)
	}
	json.Unmarshal([]byte(result), &assignment)
	
	assignment.Progress = newVal
	str, _ := json.Marshal(assignment)
	err = d.rdb.HSet(d.ctx, d.key("assignments"), assignmentId, string(str)).Err()
	if err != nil {
		return wd.NewError(wd.WorkloadAssignmentErr, err, d.name, d.workload.Name, assignmentId)
	}

	return nil
}

func (d *RedisDistributor) Commit(assignmentId string) error {
	lockVal, err := d.acquireLock("assignments")
	if err != nil {
		return wd.NewError(wd.LockErr, err, d.name, "", "")
	}
	defer d.releaseLock(lockVal, "assignments")

	err = d.rdb.HDel(d.ctx, d.key("assignments"), assignmentId).Err()
	if err != nil && err != redis.Nil {
		return wd.NewError(wd.WorkloadAssignmentErr, err, d.name, d.workload.Name, assignmentId)
	}

	return nil
}

func (d *RedisDistributor) RollBack(assignmentId string) error {
	// Acquire the locks
	lockVal, err := d.acquireLock("assignment_index", "available", "assignments")
	if err != nil {
		return wd.NewError(wd.LockErr, err, d.name, "", "")
	}
	defer d.releaseLock(lockVal, "assignment_index", "available", "assignments")

	var assignment wd.Assignment
	str, err := d.rdb.HGet(d.ctx, d.key("assignments"), assignmentId).Result()
	if err != nil && err != redis.Nil {
		return wd.NewError(wd.WorkloadAssignmentErr, err, d.name, d.workload.Name, assignmentId)
	}
	json.Unmarshal([]byte(str), &assignment)
	// Mark this batch "failed" as its rollbacked
	// next time the client retrieve the batch
	// check the batch "HasFailed" and "Progress"
	// to continue handle the batch at where it interupted
	assignment.HasFailed = true
	assignmentJson, _ := json.Marshal(assignment)
	err = d.rdb.HSet(d.ctx, d.key("assignments"), assignmentId, string(assignmentJson)).Err()
	if err != nil && err != redis.Nil {
		return wd.NewError(wd.WorkloadAssignmentErr, err, d.name, d.workload.Name, assignmentId)
	}
	// Set the "available" with the batch data at "index - 1"
	// and future call to Next() will return this batch
	idx, err := d.rdb.Get(d.ctx, d.key("assignment_index")).Int()
	if err != nil && err != redis.Nil {
		return wd.NewError(wd.WorkloadAssignmentErr, err, d.name, d.workload.Name, assignmentId)
	}
	available := fmt.Sprintf("[%v, %v]", assignment.Start, assignment.End)
	err = d.rdb.HSet(d.ctx, d.key("available"), strconv.Itoa(idx-1), available).Err()
	if err != nil {
		return wd.NewError(wd.WorkloadAssignmentErr, err, d.name, d.workload.Name, assignmentId)
	}
	// Update the index value
	err = d.rdb.Decr(d.ctx, d.key("assignment_index")).Err()
	if err != nil {
		return wd.NewError(wd.WorkloadAssignmentErr, err, d.name, d.workload.Name, assignmentId)
	}
	return nil
}

func (d *RedisDistributor) key(k string) string {
	switch k {
	case "workloads":
		return fmt.Sprintf("work_distributor:%v:workloads", d.name)
	default:
		return fmt.Sprintf("work_distributor:%v:workload:%v:%v", d.name, d.workload.Name, k)
	}
}

func (d *RedisDistributor) lock(k string) string {
	switch k {
	case "workloads":
		return fmt.Sprintf("lock:work_distributor:%v:workloads", d.name)
	default:
		return fmt.Sprintf("lock:work_distributor:%v:workload:%v:%v", d.name, d.workload.Name, k)
	}
}

func (d *RedisDistributor) available() ([2]int, *wd.Error) {
	lockVal, err := d.acquireLock("assignment_index", "available")
	if err != nil {
		return [2]int{}, wd.NewError(wd.LockErr, err, d.name, "", "")
	}
	defer d.releaseLock(lockVal, "assignment_index", "available")

	var result [2]int
	// Get index
	idx, err := d.rdb.Get(d.ctx, d.key("assignment_index")).Int()
	if err != nil && err != redis.Nil {
		return [2]int{}, wd.NewError(wd.WorkloadAssignmentErr, err, d.name, d.workload.Name, "")
	}
	// Nil batch
	if idx == d.workload.NumOfAssignments() {
		return [2]int{}, wd.NewError(wd.WorkloadEmpty, nil, d.name, d.workload.Name, "")
	}
	// Get next assignment unit start, and end from "available"
	exists, err := d.rdb.HExists(d.ctx, d.key("available"), strconv.Itoa(idx)).Result()
	if err != nil {
		return [2]int{}, wd.NewError(wd.WorkloadAssignmentErr, err, d.name, d.workload.Name, "")
	}
	// Create new available assignment if not exists
	if !exists {
		s := idx * d.workload.DistributionSize + 1
		e := s + d.workload.DistributionSize - 1
		if e > d.workload.NumOfUnits {
			e = d.workload.NumOfUnits
		}
		str := fmt.Sprintf("[%v, %v]", s, e)
		err := d.rdb.HSet(d.ctx, d.key("available"), strconv.Itoa(idx), str).Err()
		if err != nil {
			return [2]int{}, wd.NewError(wd.WorkloadAssignmentErr, err, d.name, d.workload.Name, "")
		}
		result = [2]int{s, e}
	} else {
		val, err := d.rdb.HGet(d.ctx, d.key("available"), strconv.Itoa(idx)).Result()
		if err != nil && err != redis.Nil {
			return [2]int{}, wd.NewError(wd.WorkloadAssignmentErr, err, d.name, d.workload.Name, "")
		}
		json.Unmarshal([]byte(val), &result)
	}

	// Increase index
	err = d.rdb.Set(d.ctx, d.key("assignment_index"), idx+1, 0).Err()
	if err != nil {
		return [2]int{}, wd.NewError(wd.WorkloadAssignmentErr, err, d.name, d.workload.Name, "")
	}

	return result, nil
}

func (d *RedisDistributor) assign(s int, e int) (*wd.Assignment, *wd.Error) {
	lockVal, err := d.acquireLock("assignments")
	if err != nil {
		return nil, wd.NewError(wd.LockErr, err, d.name, "", "")
	}
	defer d.releaseLock(lockVal, "assignments")

	var assignment wd.Assignment
	assignmentId := fmt.Sprintf("%v-%v", s, e)
	result, err := d.rdb.HGet(d.ctx, d.key("assignments"), assignmentId).Result()
	if err != nil && err != redis.Nil {
		return nil, wd.NewError(wd.WorkloadAssignmentErr, err, d.name, d.workload.Name, "")
	}
	if err == redis.Nil {
		assignment = wd.Assignment{
			Id:        assignmentId,
			Start:     s,
			End:       e,
			HasFailed: false,
			Progress:  0,
		}
		str, _ := json.Marshal(assignment)
		err := d.rdb.HSet(d.ctx, d.key("assignments"), assignmentId, string(str)).Err()
		if err != nil {
			return nil, wd.NewError(wd.WorkloadAssignmentErr, err, d.name, d.workload.Name, "")
		}
	} else {
		json.Unmarshal([]byte(result), &assignment)
	}

	return &assignment, nil
}

func (d *RedisDistributor) acquireLock(keys... string) (string, error) {
	redisKeys := make([]string, len(keys))
	for i, k := range keys {
		redisKeys[i] = d.lock(k)
	}
	lockVal := strconv.Itoa(int(time.Now().UnixMilli()))
	err := acquireLock(d.ctx, d.rdb, d.opts.LockTimeOut, lockVal, redisKeys...)
	return lockVal, err
}

func (d *RedisDistributor) releaseLock(lock string, keys... string) {
	redisKeys := make([]string, len(keys))
	for i, k := range keys {
		redisKeys[i] = d.lock(k)
	}
	releaseLock(d.ctx, d.rdb, lock, redisKeys...)
}
