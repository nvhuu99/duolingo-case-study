package query

import (
	"duolingo/constant"
	"duolingo/repository/log"
	"duolingo/repository/log/param"
	"duolingo/repository/log/result"
	"errors"
)

type WorkLoadMetricQuery struct {
	repo *log.LogRepo
	params *param.WorkloadMetricQueryParam
	results []*result.WorkloadMetricQueryResult
	reduction *param.WorkloadMetricReduction
}

func NewWorkloadMetricQuery(repo *log.LogRepo) *WorkLoadMetricQuery {
	return &WorkLoadMetricQuery{ repo: repo }
}

func (query *WorkLoadMetricQuery) SetParams(params *param.WorkloadMetricQueryParam) *WorkLoadMetricQuery {
	query.params = params
	return query
}

func (query *WorkLoadMetricQuery) SetReduction(reduction *param.WorkloadMetricReduction) *WorkLoadMetricQuery {
	query.reduction = reduction
	return query
}

func (query *WorkLoadMetricQuery) Execute() error {
	if query.params == nil {
		return errors.New("workload metric query params missing")
	}
	results, err := query.repo.WorkloadServiceMetrics(query.params)
	if err != nil {
		return err
	}
	query.results = results
	return nil
}

func (query *WorkLoadMetricQuery) Reduce() error {
	if query.reduction == nil {
		return errors.New("workload reduction params missing")
	}

	workload, err := query.repo.GetWorkloadMetadata(query.params.Filters.TraceId)
    if err != nil {
        return err
    }

	for _, r := range query.results {
		if query.params.MetricTarget == constant.METRIC_TARGET_REDIS {
			wrapper := result.NewWorkloadRedisQueryResult(r)
			err := wrapper.Reduce(workload, query.reduction.ReductionStep, query.reduction.Stratergies)
			if err != nil {
				return err
			}
		} else if query.params.MetricTarget == constant.METRIC_TARGET_RABBITMQ {
			wrapper := result.NewWorkloadRabbitMQQueryResult(r)
			err := wrapper.Reduce(workload, query.reduction.ReductionStep, query.reduction.Stratergies)
			if err != nil {
				return err
			}
		} else if query.params.MetricTarget == constant.METRIC_TARGET_MONGO {
			wrapper := result.NewWorkloadMongoQueryResult(r)
			err := wrapper.Reduce(workload, query.reduction.ReductionStep, query.reduction.Stratergies)
			if err != nil {
				return err
			}
		} else if query.params.MetricTarget == constant.METRIC_TARGET_FIREBASE {
			wrapper := result.NewWorkloadFirebaseQueryResult(r)
			err := wrapper.Reduce(workload, query.reduction.ReductionStep, query.reduction.Stratergies)
			if err != nil {
				return err
			}
		} else {
			err := r.Reduce(workload, query.reduction.ReductionStep, query.reduction.Stratergies)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (query *WorkLoadMetricQuery) Result() []*result.WorkloadMetricQueryResult {
	return query.results
}