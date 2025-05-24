package log

import (
	"context"
	"duolingo/lib/log"
	"duolingo/repository/log/param"
	"duolingo/repository/log/result"
	"errors"
	"fmt"
	"net/url"
	"time"

	b "go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	connectionTimeOut = 10 * time.Second
	defaultTimeOut    = 30 * time.Second
)

type LogRepo struct {
	uri      string
	database string
	client   *mongo.Client
	ctx      context.Context
}

func NewLogRepo(ctx context.Context, database string) *LogRepo {
	repo := LogRepo{}
	repo.ctx = ctx
	repo.database = database

	return &repo
}

func (repo *LogRepo) SetConnection(host string, port string, usr string, pwd string) error {
	repo.uri = fmt.Sprintf(
		"mongodb://%v:%v@%v:%v/",
		url.QueryEscape(usr),
		url.QueryEscape(pwd),
		host,
		port,
	)

	opts := options.Client()
	opts.SetConnectTimeout(connectionTimeOut)
	opts.SetSocketTimeout(defaultTimeOut)
	opts.ApplyURI(repo.uri)

	client, err := mongo.Connect(repo.ctx, opts)
	if err != nil {
		return err
	}
	repo.client = client

	return nil
}

func (repo *LogRepo) GetWorkloadMetadata(traceId string) (*result.WorkloadMetadataResult, error) {
    collection := repo.client.Database(repo.database).Collection("log")
    pipeline := mongo.Pipeline{
        {{"$match", b.D{
            {"level", 1},
            {"context.trace.trace_id", traceId},
        }}},
        {{"$group", b.D{
            {"_id", b.D{
                {"service_name", "$context.trace.service_name"},
                {"trace_id", "$context.trace.trace_id"},
            }},
            {"instance_ids", b.D{{"$addToSet", "$context.trace.instance_id"}}},
            {"start_times", b.D{{"$push", "$context.trace.start_time"}}}, // temp array to test start_time
        }}},
        {{"$group", b.D{
            {"_id", nil},
            {"trace_id", b.D{{"$first", "$_id.trace_id"}}},
            {"start_time", b.D{{"$min", b.D{
                {"$toDate", b.D{{"$arrayElemAt", b.A{"$start_times", 0}}}},
            }}}},
            {"service_instances", b.D{{"$push", b.D{
                {"service_name", "$_id.service_name"},
                {"instance_ids", "$instance_ids"},
            }}}},
        }}},
        {{"$project", b.D{
            {"_id", 0},
            {"trace_id", 1},
            {"start_time", 1},
            {"service_instances", 1},
        }}},
    }

    cursor, err := collection.Aggregate(repo.ctx, pipeline)
    if err != nil {
        return nil, err
    }
    defer cursor.Close(repo.ctx)
    
    result := new(result.WorkloadMetadataResult)
    if cursor.Next(repo.ctx) {
        if err := cursor.Decode(result); err != nil {
            return nil, err
        }
        return result, nil
    }
    return nil, fmt.Errorf("workload not found (trace_id: \"%v\")", traceId)
}

func (repo *LogRepo) ListWorkloadOperations(traceId string) ([]*result.WorkloadOperationListResult, error) {
    pipeline := mongo.Pipeline{
        b.D{{"$match", b.D{
            {"level", log.LevelInfo},
            {"context.trace.trace_id", traceId},
        }}},
        b.D{{"$project", b.D{
            {"service_name", "$context.trace.service_name"},
            {"service_operation", "$context.trace.service_operation"},
            {"start_time", b.D{{"$toDate", "$context.trace.start_time"}}},
        }}},
        b.D{{"$group", b.D{
            {"_id", b.D{
                {"service_name", "$service_name"},
                {"service_operation", "$service_operation"},
            }},
            {"start_time", b.D{{"$min", "$start_time"}}},
        }}},
        b.D{{"$project", b.D{
            {"_id", 0},
            {"service_name", "$_id.service_name"},
            {"service_operation", "$_id.service_operation"},
            {"start_time", "$start_time"},
        }}},
        b.D{{"$sort", b.D{
            {"start_time", 1},
        }}},
    }
    collection := repo.client.Database(repo.database).Collection("log")
    cursor, err := collection.Aggregate(repo.ctx, pipeline)
    if err != nil {
        return nil, err
    }
    defer cursor.Close(repo.ctx)

    response := []*result.WorkloadOperationListResult{}
    for cursor.Next(repo.ctx) {
        aggr := new(result.WorkloadOperationListResult)
        if err := cursor.Decode(aggr); err != nil {
            return nil, err
        }
        response = append(response, aggr)
    }
    
    return response, nil
}

func (repo *LogRepo) GetWorkloadOptsExecTimeSpans(traceId string) ([]*result.OperationExecTimeSpanResult, error) {
    workload, err := repo.GetWorkloadMetadata(traceId)
    if err != nil {
        return nil, err
    }
    pipeline := mongo.Pipeline{
        b.D{{"$match", b.D{
            {"level", log.LevelInfo},
            {"context.trace.trace_id", traceId},
        }}},
        b.D{{"$project", b.D{
            {"service_name", "$context.trace.service_name"},
            {"service_operation", "$context.trace.service_operation"},
            {"operation_start_latency_ms", b.D{
                {"$subtract", b.A{
                    b.D{{"$toDate", "$context.trace.start_time"}},
                    workload.StartTime,
                }},
            }},
            {"operation_end_latency_ms", b.D{
                {"$subtract", b.A{
                    b.D{{"$toDate", "$context.trace.end_time"}},
                    workload.StartTime,
                }},
            }},
        }}},
        b.D{{"$group", b.D{
            {"_id", b.D{
                {"service_name", "$service_name"},
                {"service_operation", "$service_operation"},
            }},
            {"operation_start_latency_ms", b.D{{"$min", "$operation_start_latency_ms"}}},
            {"operation_end_latency_ms", b.D{{"$max", "$operation_end_latency_ms"}}},
        }}},
        b.D{{"$project", b.D{
            {"_id", 0},
            {"service_name", "$_id.service_name"},
            {"service_operation", "$_id.service_operation"},
            {"operation_start_latency_ms", "$operation_start_latency_ms"},
            {"operation_end_latency_ms", "$operation_end_latency_ms"},
        }}},
    }
    collection := repo.client.Database(repo.database).Collection("log")
    cursor, err := collection.Aggregate(repo.ctx, pipeline)
    if err != nil {
        return nil, err
    }
    defer cursor.Close(repo.ctx)

    response := []*result.OperationExecTimeSpanResult{}
    for cursor.Next(repo.ctx) {
        aggr := new(result.OperationExecTimeSpanResult)
        if err := cursor.Decode(aggr); err != nil {
            return nil, err
        }
        response = append(response, aggr)
    }
    
    return response, nil
}

func (repo *LogRepo) WorkloadServiceMetrics(params *param.WorkloadMetricQuery, downsampling *param.WorkloadMetricDownsampling) ([]*result.WorkloadMetricQueryResult, error) {
    workload, err := repo.GetWorkloadMetadata(params.Filters.TraceId)
    if err != nil {
        return nil, err
    }

    filters, err := repo.filtersForWorkloadServiceMetric(params)
    if err != nil {
        return nil, err
    }

    pipeline := mongo.Pipeline{
        {{"$unwind", b.D{{"path", "$data.metric.snapshots"}}}},
        {{"$match", filters}},
		{{"$group", b.D{
			{"_id", b.D{
				{"metric_target", "$data.metric.tags.metric_target"},
				{"metric_name", "$data.metric.tags.metric_name"},
			}},
			{"snapshots", b.D{
				{"$push", b.D{
					{"timestamp", "$data.metric.snapshots.timestamp"},
					{"value", "$data.metric.snapshots.value"},
				}},
			}},
		}}},
		{{"$project", b.D{
            {"_id", 0},
            {"metric_target", "$_id.metric_target"},
            {"metric_name", "$_id.metric_name"},
            {"snapshots", "$snapshots"},
        }}},
	}

	collection := repo.client.Database(repo.database).Collection("log")
    cursor, err := collection.Aggregate(repo.ctx, pipeline)
    if err != nil {
        return nil, err
    }
    defer cursor.Close(repo.ctx)

    response := []*result.WorkloadMetricQueryResult{}
    for cursor.Next(repo.ctx) {
        stats := new(result.WorkloadMetricQueryResult)
        if err := cursor.Decode(stats); err != nil {
            return nil, err
        }
        if err := stats.Downsampling(workload.StartTime, downsampling.ReductionStep, downsampling.Stratergies); err != nil {
            return nil, err
        }
        response = append(response, stats)
    }
    
    return response, nil
}

func (repo *LogRepo) filtersForWorkloadServiceMetric(params *param.WorkloadMetricQuery) (b.D, error) {
    if params.Filters.TraceId == "" {
        return nil, errors.New("workload service metric trace id is empty string")
    }

    filters := b.D{
        {"level", log.LevelDebug},
        {"context.trace.trace_id", params.Filters.TraceId },
    }
    if params.Filters.ServiceName != "" {
        filters = append(filters, b.E{"context.trace.service_name", params.Filters.ServiceName})
    }
    if params.Filters.ServiceOperation != "" {
        filters = append(filters, b.E{"context.trace.service_operation", params.Filters.ServiceOperation})
    }
    instancesIds := []string{}
    for _, id := range params.Filters.InstanceIds {
        if id != "" {
            instancesIds = append(instancesIds, id)
        }
    }
    if len(instancesIds) > 0 {
        filters = append(filters, b.E{"context.trace.instance_id", b.D{{"$in", instancesIds}}})
    }
    if len(params.MetricGroups) == 0 {
        return nil, errors.New("workload service metric query metric groups not specified")
    }
    targetFilters := b.A{}
    for _, grp := range params.MetricGroups {
        condition := b.D{}
        if grp.MetricTarget != "" {
            condition = append(condition, b.E{"data.metric.tags.metric_target", grp.MetricTarget})
        }
        if grp.MetricName != "" {
            condition = append(condition, b.E{"data.metric.tags.metric_name", grp.MetricName})
        }
        if len(condition) > 0 {
            targetFilters = append(targetFilters, condition)
        }
    }
    filters = append(filters, b.E{"$or", targetFilters})

    return filters, nil
}

func (repo *LogRepo) WorkloadServiceMetricSummary(params *param.WorkloadMetricQuery) ([]*result.WorkloadMetricSummaryResult, error) {
    filters, err := repo.filtersForWorkloadServiceMetric(params)
    if err != nil {
        return nil, err
    }

    pipeline := mongo.Pipeline{
        {{"$unwind", b.D{{"path", "$data.metric.snapshots"}}}},
        {{"$match", filters}},
        {{"$group", b.D{
			{"_id", b.D{
				{"metric_target", "$data.metric.tags.metric_target"},
				{"metric_name", "$data.metric.tags.metric_name"},
			}},
			{"values", b.D{{"$push", "$data.metric.snapshots.value"}}},
            {"min_value", b.D{{"$min", "$data.metric.snapshots.value"}}},
            {"max_value", b.D{{"$max", "$data.metric.snapshots.value"}}},
            {"avg_value", b.D{{"$avg", "$data.metric.snapshots.value"}}},
		}}},
        {{"$addFields", b.D{
            {"sorted_values", b.D{
                {"$sortArray", b.D{
                    {"input", "$values"},
                    {"sortBy", 1},
                }},
            }},
        }}},
        {{"$addFields", b.D{
            {"median", b.D{{"$let", b.D{
                {"vars", b.D{
                    {"half", b.D{{"$divide", b.A{b.D{{"$size", "$sorted_values"}}, 2}}}},
                }},
                {"in", b.D{
                    {"$cond", b.A{
                        b.D{{"$eq", b.A{b.D{{"$mod", b.A{"$$half", 1}}}, 0}}},
                        b.D{{"$avg", b.A{
                            b.D{{"$arrayElemAt", b.A{"$sorted_values", b.D{{"$subtract", b.A{"$$half", 1}}}}}},
                            b.D{{"$arrayElemAt", b.A{"$sorted_values", "$$half"}}},
                        }}},
                        b.D{{"$arrayElemAt", b.A{"$sorted_values", b.D{{"$floor", "$$half"}}}}}, // Added comma here
                    }},
                }},
            }}}},
            {"p5", b.D{{"$arrayElemAt", b.A{
                "$sorted_values",
                b.D{{"$floor", b.D{{"$multiply", b.A{0.05, b.D{{"$subtract", b.A{b.D{{"$size", "$sorted_values"}}, 1}}}}}}}},
            }}}},
            {"p25", b.D{{"$arrayElemAt", b.A{
                "$sorted_values",
                b.D{{"$floor", b.D{{"$multiply", b.A{0.25, b.D{{"$subtract", b.A{b.D{{"$size", "$sorted_values"}}, 1}}}}}}}},
            }}}},
            {"p75", b.D{{"$arrayElemAt", b.A{
                "$sorted_values",
                b.D{{"$floor", b.D{{"$multiply", b.A{0.75, b.D{{"$subtract", b.A{b.D{{"$size", "$sorted_values"}}, 1}}}}}}}},
            }}}},
            {"p95", b.D{{"$arrayElemAt", b.A{
                "$sorted_values",
                b.D{{"$floor", b.D{{"$multiply", b.A{0.95, b.D{{"$subtract", b.A{b.D{{"$size", "$sorted_values"}}, 1}}}}}}}},
            }}}},
        }}},
        {{"$project", b.D{
            {"_id", 0},
            {"metric_target", "$_id.metric_target"},
            {"metric_name", "$_id.metric_name"},
            {"values", "$sorted_values"},
            {"summary", b.D{
                {"minimum", "$min_value"},
                {"maximum", "$max_value"},
                {"average", "$avg_value"},
                {"median", "$median"},
                {"p5", "$p5"},
                {"p25", "$p25"},
                {"p50", "$p50"},
                {"p75", "$p75"},
                {"p95", "$p95"},
            }},
        }}},
    }
    collection := repo.client.Database(repo.database).Collection("log")
    cursor, err := collection.Aggregate(repo.ctx, pipeline)
    if err != nil {
        return nil, err
    }
    defer cursor.Close(repo.ctx)

    response := []*result.WorkloadMetricSummaryResult{}
    for cursor.Next(repo.ctx) {
        aggr := new(result.WorkloadMetricSummaryResult)
        if err := cursor.Decode(aggr); err != nil {
            return nil, err
        }
        response = append(response, aggr)
    }
    
    return response, nil
}