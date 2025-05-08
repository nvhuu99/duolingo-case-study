package main

import (
	rest "duolingo/lib/rest_http"
	sv "duolingo/lib/service_container"
	"duolingo/service/metric_api/bootstrap"

	log_repo "duolingo/repository/log"
	"log"
)

var (
	container *sv.ServiceContainer
	repo *log_repo.LogRepo
	server    *rest.Server
)

func aggregateWorkloadOptsExecTime(request *rest.Request, response *rest.Response) {
	traceId := request.Path("traceId").Str()
	if traceId == "" {
		response.InvalidRequest("", map[string]string{
			"traceId": "traceId must not be empty",
		})
		return
	}

	report, err := repo.GetWorkloadOptsExecTimeSpans(traceId)
	if err != nil {
		response.ServerErr("", err)
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
		response.ServerErr("", err)
		return
	}

	response.Ok("", report)
}

func main() {
	bootstrap.Run()

	container = sv.GetContainer()
	server = container.Resolve("rest.server").(*rest.Server)
	repo = container.Resolve("repo.log").(*log_repo.LogRepo)

	server.Router().Get("/metric/workload/{traceId}/operations", listWorkloadOperations)
	server.Router().Get("/metric/workload/{traceId}/operations/report-execution-time-spans", aggregateWorkloadOptsExecTime)

	log.Println("serving metric api")

	server.Serve()
}
