package main

import (
	"duolingo/lib/metric/reduction"
	rest "duolingo/lib/rest_http"
	sv "duolingo/lib/service_container"
	"duolingo/service/metric_api/bootstrap"
	// cnst "duolingo/constant"

	log_repo "duolingo/repository/log"
	"duolingo/repository/log/param"
	"duolingo/repository/log/query"
	"log"
)

var (
	container *sv.ServiceContainer
	repo      *log_repo.LogRepo
	server    *rest.Server
	strategies map[string]reduction.ReductionStrategy
)

func serviceExecutionTimeSpans(request *rest.Request, response *rest.Response) {
	traceId := request.Path("traceId").Str()
	if traceId == "" {
		response.InvalidRequest("", map[string]string{
			"traceId": "traceId must not be empty",
		})
		return
	}

	report, err := repo.GetWorkloadOptsExecTimeSpans(traceId)
	if err != nil {
		response.ServerErr("", err.Error())
		log.Println("service exec time span", err)
		return
	}

	response.Ok("", report)
}

func listWorkloadOperations(request *rest.Request, response *rest.Response) {
	traceId := request.Path("traceId").Str()
	if traceId == "" {
		response.InvalidRequest("", map[string]string{
			"traceId": "traceId must not be empty",
		})
		return
	}

	report, err := repo.ListWorkloadOperations(traceId)
	if err != nil {
		response.ServerErr("", err.Error())
		log.Println("list workload operation", err)
		return
	}

	response.Ok("", report)
}

func workloadMetadata(request *rest.Request, response *rest.Response) {
	traceId := request.Path("traceId").Str()
	if traceId == "" {
		response.InvalidRequest("", map[string]string{
			"traceId": "traceId must not be empty",
		})
		return
	}

	metadata, err := repo.GetWorkloadMetadata(traceId)
	if err != nil {
		response.ServerErr("", err.Error())
		log.Println("get workload metadata:", err)
		return
	}

	response.Ok("", metadata)
}

func serviceMetrics(request *rest.Request, response *rest.Response) {
	traceId := request.Path("traceId").Str()
	if traceId == "" {
		response.InvalidRequest("", map[string]string{"traceId": "traceId must not be empty"})
		return
	}

	serviceName := request.Input("service_name").Str()
	serviceOperation := request.Input("service_operation").Str()
	instanceIds, _ := request.Input("instance_ids").StrArr()
	metricNames, _ := request.Input("metric_names").StrArr()
	reductionStep, _ := request.Input("reduction_step").Int64()
	strgs, _ := request.Input("strategies").StrArr()
	
	reductionStrategies := make(map[string]reduction.ReductionStrategy)
	for _, name := range strgs {
		if _, exist := strategies[name]; !exist {
			response.InvalidRequest("", map[string]string{"strategies": name + " is not a supported reduction strategy"})
			return
		}
		reductionStrategies[name] = strategies[name]
	}

	reduction := &param.WorkloadMetricReduction{ ReductionStep: reductionStep, Stratergies: reductionStrategies }
	params := param.NewWorkloadMetricQueryParam(traceId).
		SetServiceName(serviceName).
		SetServiceOperation(serviceOperation).
		SetServiceInstanceIds(instanceIds).
		SetMetricTarget(serviceName).
		AddMetricNames(metricNames...)

	var err error
	var data any
	q := query.NewWorkloadMetricQuery(repo).SetParams(params).SetReduction(reduction)
	if err = q.Execute(); err == nil {
		if err = q.Reduce(); err == nil {
			data = q.Result()
		}
	}

	if err != nil {
		response.ServerErr("", err.Error())
		log.Println("workload service metrics", err)
		return
	}

	response.Ok("", data)
}

func redisMetrics(request *rest.Request, response *rest.Response) {
	traceId := request.Path("traceId").Str()
	if traceId == "" {
		response.InvalidRequest("", map[string]string{"traceId": "traceId must not be empty"})
		return
	}

	metricNames, _ := request.Input("metric_names").StrArr()
	reductionStep, _ := request.Input("reduction_step").Int64()

	reduction := &param.WorkloadMetricReduction{ ReductionStep: reductionStep, Stratergies: strategies }
	params := param.NewWorkloadRedisMetricQuery(traceId).AddMetricNames(metricNames...)

	var err error
	var data any
	q := query.NewWorkloadMetricQuery(repo).SetParams(params.GetQuery()).SetReduction(reduction)
	if err = q.Execute(); err == nil {
		if err = q.Reduce(); err == nil {
			data = q.Result()
		}
	}

	if err != nil {
		response.ServerErr("", err.Error())
		log.Println("workload redis metrics", err)
		return
	}

	response.Ok("", data)
}

func rabbitMQMetrics(request *rest.Request, response *rest.Response) {
	traceId := request.Path("traceId").Str()
	if traceId == "" {
		response.InvalidRequest("", map[string]string{"traceId": "traceId must not be empty"})
		return
	}

	metricNames, _ := request.Input("metric_names").StrArr()
	reductionStep, _ := request.Input("reduction_step").Int64()
	queue := request.Input("queue").Str()

	reduction := &param.WorkloadMetricReduction{ ReductionStep: reductionStep, Stratergies: strategies }
	params := param.NewWorkloadRabbitMQMetricQuery(traceId).
				AddMetricNames(metricNames...).
				SetMessageQueue(queue)

	var err error
	var data any
	q := query.NewWorkloadMetricQuery(repo).SetParams(params.GetQuery()).SetReduction(reduction)
	if err = q.Execute(); err == nil {
		if err = q.Reduce(); err == nil {
			data = q.Result()
		}
	}

	if err != nil {
		response.ServerErr("", err.Error())
		log.Println("workload rabbitmq metrics", err)
		return
	}

	response.Ok("", data)
}

func mongoMetrics(request *rest.Request, response *rest.Response) {
	traceId := request.Path("traceId").Str()
	if traceId == "" {
		response.InvalidRequest("", map[string]string{"traceId": "traceId must not be empty"})
		return
	}

	metricNames, _ := request.Input("metric_names").StrArr()
	reductionStep, _ := request.Input("reduction_step").Int64()

	reduction := &param.WorkloadMetricReduction{ ReductionStep: reductionStep, Stratergies: strategies }
	params := param.NewWorkloadMongoMetricQuery(traceId).AddMetricNames(metricNames...)

	var err error
	var data any
	q := query.NewWorkloadMetricQuery(repo).SetParams(params.GetQuery()).SetReduction(reduction)
	if err = q.Execute(); err == nil {
		if err = q.Reduce(); err == nil {
			data = q.Result()
		}
	}

	if err != nil {
		response.ServerErr("", err.Error())
		log.Println("workload mongo metrics", err)
		return
	}

	response.Ok("", data)
}

func firebaseMetrics(request *rest.Request, response *rest.Response) {
	traceId := request.Path("traceId").Str()
	if traceId == "" {
		response.InvalidRequest("", map[string]string{"traceId": "traceId must not be empty"})
		return
	}

	metricNames, _ := request.Input("metric_names").StrArr()
	reductionStep, _ := request.Input("reduction_step").Int64()

	reduction := &param.WorkloadMetricReduction{ ReductionStep: reductionStep, Stratergies: strategies }
	params := param.NewWorkloadFirebaseMetricQuery(traceId).AddMetricNames(metricNames...)

	var err error
	var data any
	q := query.NewWorkloadMetricQuery(repo).SetParams(params.GetQuery()).SetReduction(reduction)
	if err = q.Execute(); err == nil {
		if err = q.Reduce(); err == nil {
			data = q.Result()
		}
	}

	if err != nil {
		response.ServerErr("", err.Error())
		log.Println("workload firebase metrics", err)
		return
	}

	response.Ok("", data)
}

func main() {
	bootstrap.Run()

	container = sv.GetContainer()
	server = container.Resolve("rest.server").(*rest.Server)
	repo = container.Resolve("repo.log").(*log_repo.LogRepo)
	strategies = map[string]reduction.ReductionStrategy{
		"median": new(reduction.Median),
		"lttb":   new(reduction.LTTB),
		"min":   new(reduction.Min),
		"max":   new(reduction.Max),
		"p1":     reduction.NewPercentileStrategy(1),
		"p5":     reduction.NewPercentileStrategy(5),
		"p25":    reduction.NewPercentileStrategy(25),
		"p75":    reduction.NewPercentileStrategy(75),
		"p95":    reduction.NewPercentileStrategy(95),
		"p99":    reduction.NewPercentileStrategy(99),
	}

	server.Router().Get("/metric/workload/{traceId}/list-operations", listWorkloadOperations)
	server.Router().Get("/metric/workload/{traceId}/service-execution-time-spans", serviceExecutionTimeSpans)
	server.Router().Get("/metric/workload/{traceId}/workload-metadata", workloadMetadata)
	server.Router().Post("/metric/workload/{traceId}/service-metrics", serviceMetrics)
	server.Router().Post("/metric/workload/{traceId}/redis-metrics", redisMetrics)
	server.Router().Post("/metric/workload/{traceId}/rabbitmq-metrics", rabbitMQMetrics)
	server.Router().Post("/metric/workload/{traceId}/mongo-metrics", mongoMetrics)
	server.Router().Post("/metric/workload/{traceId}/firebase-metrics", firebaseMetrics)
	server.Router().Options("*", func(request *rest.Request, response *rest.Response) {
		response.Ok("", nil)
	})

	log.Println("serving metric api")

	server.Serve()
}
