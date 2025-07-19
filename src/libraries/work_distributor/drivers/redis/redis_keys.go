package redis

import (
	"duolingo/libraries/work_distributor"
	"encoding/json"

	"github.com/redis/go-redis/v9"
)

func workloadKey(workloadId string) string {
	return "work_distributor:workload:" + workloadId
}

func assignmentsOfWorkloadKey(workloadId string) string {
	return "work_distributor:workload_assignments:" + workloadId
}

func errOrRedisNilAlias(err error, ifNilErr error) error {
	if err == redis.Nil {
		return ifNilErr
	}
	return err
}

func unmarshalWorkloadIgnoreErr(jsonStr string) *work_distributor.Workload {
	workload := new(work_distributor.Workload)
	json.Unmarshal([]byte(jsonStr), workload)
	return workload
}

func marshalWorkloadIgnoreErr(w *work_distributor.Workload) string {
	marshaled, _ := json.Marshal(w)
	return string(marshaled)
}
