package main

import (
	"duolingo/lib/metric/downsampling"
	rest "duolingo/lib/rest_http"
	sv "duolingo/lib/service_container"
	"duolingo/service/metric_api/bootstrap"

	log_repo "duolingo/repository/log"
	"duolingo/repository/log/param"
	"log"
)

var (
	container *sv.ServiceContainer
	repo *log_repo.LogRepo
	server    *rest.Server
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

func serviceMetrics(request *rest.Request, response *rest.Response) {
	traceId := request.Path("traceId").Str()
	if traceId == "" {
		response.InvalidRequest("", map[string]string{
			"traceId": "traceId must not be empty",
		})
		return
	}

	serviceName := request.Input("service_name").Str()
	serviceOperation := request.Input("service_operation").Str()
	instanceIds, _ := request.Input("instance_ids").StrArr()
	metricNames, _ := request.Input("metric_names").StrArr()
	reductionStep, _ := request.Input("reduction_step").Int64()
	summary, _ := request.Input("summary").Bool()

	strategies := map[string]downsampling.DownsamplingStrategy{
		"median": new(downsampling.MovingAverage),
		"lttb": new(downsampling.LTTB),
		"p1": downsampling.NewPercentileStrategy(1),
		"p5": downsampling.NewPercentileStrategy(5),
		"p25": downsampling.NewPercentileStrategy(25),
		"p75": downsampling.NewPercentileStrategy(75),
		"p95": downsampling.NewPercentileStrategy(95),
		"p99": downsampling.NewPercentileStrategy(99),
	}
	reduction := &param.WorkloadMetricDownsampling{
		ReductionStep: reductionStep,
		Stratergies: strategies,
	}
	query := param.WorkloadMetricQueryParams(traceId).
				SetServiceName(serviceName).
				SetServiceOperation(serviceOperation).
				SetServiceInstanceIds(instanceIds)
	for _, metricName := range metricNames {
		query.AddMetricGroup(serviceName, metricName)
	}

	var report any
	var err error
	if summary {
		report, err = repo.WorkloadServiceMetricSummary(query)
	} else {
		report, err = repo.WorkloadServiceMetrics(query, reduction)
	}
	if err != nil {
		response.ServerErr("", err.Error())
		log.Println("workload service metrics", err)
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


func main() {
	bootstrap.Run()

	container = sv.GetContainer()
	server = container.Resolve("rest.server").(*rest.Server)
	repo = container.Resolve("repo.log").(*log_repo.LogRepo)

	server.Router().Get("/metric/workload/{traceId}/list-operations", listWorkloadOperations)
	server.Router().Get("/metric/workload/{traceId}/service-execution-time-spans", serviceExecutionTimeSpans)
	server.Router().Get("/metric/workload/{traceId}/workload-metadata", workloadMetadata)
	server.Router().Post("/metric/workload/{traceId}/service-metrics", serviceMetrics)
	server.Router().Options("*", func (request *rest.Request, response *rest.Response) {
		response.Ok("", nil)
	})


	log.Println("serving metric api")

	server.Serve()
}
