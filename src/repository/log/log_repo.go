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
            {"level", b.D{{"$in", b.A{log.LevelInfo, log.LevelDebug}}}},
            {"context.trace.trace_id", traceId},
        }}},
        {{"$group", b.D{
            {"_id", b.D{
                {"service_name", "$context.trace.service_name"},
                {"trace_id", "$context.trace.trace_id"},
            }},
            {"incr_ms", b.D{{"$max", "$data.metric.incr_ms"}}},
            {"instance_ids", b.D{{"$addToSet", "$context.trace.instance_id"}}},
            {"start_time", b.D{{"$min", b.D{{"$toDate", "$context.trace.start_time"}}}}},
            {"end_time", b.D{{"$max", b.D{{"$toDate", "$context.trace.end_time"}}}}},
        }}},
        {{"$group", b.D{
            {"_id", nil},
            {"trace_id", b.D{{"$first", "$_id.trace_id"}}},
            {"incr_ms", b.D{{"$max", "$incr_ms"}}},
            {"start_time", b.D{{"$min", "$start_time"}}},
            {"end_time", b.D{{"$max", "$end_time"}}},
            {"service_instances", b.D{{"$push", b.D{
                {"service_name", "$_id.service_name"},
                {"instance_ids", "$instance_ids"},
            }}}},
        }}},
        {{"$project", b.D{
            {"_id", 0},
            {"trace_id", 1},
            {"incr_ms", 1},
            {"start_time", 1},
            {"end_time", 1},
            {"duration_ms", b.D{{"$subtract", b.A{"$end_time","$start_time"}}}},
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

func (repo *LogRepo) WorkloadServiceMetrics(params *param.WorkloadMetricQueryParam) ([]*result.WorkloadMetricQueryResult, error) {
    filters, err := repo.filtersForWorkloadServiceMetric(params)
    if err != nil {
        return nil, err
    }

    workload, err := repo.GetWorkloadMetadata(params.Filters.TraceId)
    if err != nil {
        return nil, err
    }

    workloadStart := workload.StartTime
    pipeline := mongo.Pipeline{
        // 1) unwind and filter
        {{"$unwind", b.D{{"path", "$data.metric.snapshots"}}}},
        {{"$match", filters}},
        
        {{"$addFields", b.D{
            {"start_time_offset", b.D{
                {"$multiply", b.A{
                    "$data.metric.incr_ms",
                    b.D{{"$ceil", b.D{
                        {"$divide", b.A{
                            b.D{{"$subtract", b.A{b.D{{"$toDate", "$data.metric.snapshots.timestamp"}}, workloadStart}}},
                            "$data.metric.incr_ms",
                        }},
                    }}},
                }},
            }},
        }}},

        // 3) facet out aggregate and pass-through series
        {{"$facet", b.D{
            // sum-accumulate
            {"aggregated_accumulate", mongo.Pipeline{
                {{"$match", b.D{
                    {"data.metric.snapshots.metadata.should_aggregate", b.D{{"$exists", true}, {"$ne", nil}}},
                    {"data.metric.snapshots.metadata.aggregation_accumulate", b.D{{"$exists", true}, {"$ne", nil}}},
                }}},
                {{"$group", b.D{
                    {"_id", b.D{
                        {"start_time_offset", "$start_time_offset"},
                        {"metric_target", "$data.metric.tags.metric_target"}, 
                        {"metric_name", "$data.metric.tags.metric_name"}, 
                        {"should_compute_rate", "$data.metric.snapshots.metadata.should_compute_rate"},
                    }},
                    {"value", b.D{{"$sum", "$data.metric.snapshots.value"}}},
                }}},
                {{"$project", b.D{
                    {"_id", 0},
                    {"start_time_offset", "$_id.start_time_offset"},
                    {"metric_target", "$_id.metric_target"},
                    {"metric_name", "$_id.metric_name"},
                    {"should_compute_rate", "$_id.should_compute_rate"},
                    {"timestamp", b.D{{"$add", b.A{workloadStart, "$_id.start_time_offset"}}}},
                    {"value", "$value"},
                }}},
            }},
            // max-aggregation
            {"aggregated_maximum", mongo.Pipeline{
                {{"$match", b.D{
                    {"data.metric.snapshots.metadata.should_aggregate", b.D{{"$exists", true}, {"$ne", nil}}},
                    {"data.metric.snapshots.metadata.aggregation_maximum", b.D{{"$exists", true}, {"$ne", nil}}},
                }}},
                {{"$group", b.D{
                    {"_id", b.D{
                        {"start_time_offset", "$start_time_offset"},
                        {"metric_target", "$data.metric.tags.metric_target"},
                        {"metric_name", "$data.metric.tags.metric_name"},
                        {"should_compute_rate", "$data.metric.snapshots.metadata.should_compute_rate"},
                    }},
                    {"value", b.D{{"$max", "$data.metric.snapshots.value"}}},
                }}},
                {{"$project", b.D{
                    {"_id", 0},
                    {"start_time_offset", "$_id.start_time_offset"},
                    {"metric_target", "$_id.metric_target"},
                    {"metric_name", "$_id.metric_name"},
                    {"should_compute_rate", "$_id.should_compute_rate"},
                    {"timestamp", b.D{{"$add", b.A{workloadStart, "$_id.start_time_offset"}}}},
                    {"value", "$value"},
                }}},
            }},
            // pass-through for non-aggregated
            {"non_aggregated", mongo.Pipeline{
                {{"$match", b.D{
                    {"$or", b.A{
                        b.D{{"data.metric.snapshots.metadata.should_aggregate", nil}},
                        b.D{{"data.metric.snapshots.metadata.should_aggregate", b.D{{"$exists", false}}}},
                    }},
                }}},
                {{"$project", b.D{
                    {"_id", 0},
                    {"metric_target", "$data.metric.tags.metric_target"},
                    {"metric_name", "$data.metric.tags.metric_name"},
                    {"start_time_offset", "$start_time_offset"},
                    {"timestamp", "$data.metric.snapshots.timestamp"},
                    {"value", "$data.metric.snapshots.value"},
                }}},
            }},
        }}},

        // 4) merge series, sort and compute lag on merged ordered docs
        {{"$project", b.D{
            {"merged", b.D{
                {"$concatArrays", b.A{"$aggregated_accumulate", "$aggregated_maximum", "$non_aggregated"}},
            }},
        }}},
        {{"$unwind", b.D{{"path", "$merged"}}}},
        {{"$replaceRoot", b.D{{"newRoot", "$merged"}}}},
        {{"$sort", b.D{{"metric_target", 1}, {"metric_name", 1}, {"start_time_offset", 1}}}},
        {{"$setWindowFields", b.D{
            {"partitionBy", b.D{{"metric_target", "$metric_target"}, {"metric_name", "$metric_name"}}},
            {"sortBy", b.D{{"start_time_offset", 1}}},
            {"output", b.D{
                {"prevValue", b.D{{"$shift", b.D{{"output", "$value"}, {"by", -1}}}}},
                {"prevstart_time_offset",  b.D{{"$shift", b.D{{"output", "$start_time_offset"}, {"by", -1}}}}},
            }},
        }}},

        // 5) facet rate vs final snapshots
        {{"$facet", b.D{
            {"computed_rate", mongo.Pipeline{
                {{"$match", b.D{
                    {"should_compute_rate", b.D{{"$ne", nil}}},
                    {"prevValue", b.D{{"$ne", nil}}},
                    {"prevstart_time_offset",  b.D{{"$ne", nil}}},
                }}},
                {{"$project", b.D{
                    {"_id", 0},
                    {"start_time_offset", "$start_time_offset"},
                    {"metric_target", "$metric_target"},
                    {"metric_name", "$metric_name"},
                    {"timestamp", "$timestamp"},
                    {"value", b.D{{"$divide", b.A{
                        b.D{{"$subtract", b.A{"$value", "$prevValue"}}},
                        b.D{{"$subtract", b.A{"$start_time_offset", "$prevstart_time_offset"}}},
                    }}}},
                }}},
            }},
            {"final_snapshots", mongo.Pipeline{
                {{"$match", b.D{
                    {"$or", b.A{
                        b.D{{"should_compute_rate", nil}},
                        b.D{{"should_compute_rate", b.D{{"$exists", false}}}},
                        b.D{{"prevValue", nil}},
                        b.D{{"prevstart_time_offset", nil}},
                    }},
                }}},
                {{"$project", b.D{
                    {"_id", 0},
                    {"start_time_offset", "$start_time_offset"},
                    {"metric_target", "$metric_target"},
                    {"metric_name", "$metric_name"},
                    {"timestamp", "$timestamp"},
                    {"value", "$value"},
                }}},
            }},
        }}},
        // 6) merge computed_rate and final and regroup by metric
        {{"$project", b.D{
            {"merged", b.D{
                // {"$concatArrays", b.A{"$computed_rate"}},
                {"$concatArrays", b.A{"$computed_rate", "$final_snapshots"}},
            }},
        }}},
        {{"$unwind", b.D{{"path", "$merged"}}}},
        {{"$replaceRoot", b.D{{"newRoot", "$merged"}}}},
        {{"$group", b.D{
            {"_id", b.D{
                {"metric_target", "$metric_target"},
                {"metric_name", "$metric_name"},
            }},
            {"snapshots", b.D{
                {"$push", b.D{
                    {"start_time_offset", "$start_time_offset"},
                    {"timestamp", "$timestamp"},
                    {"value", "$value"},
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

    results := []*result.WorkloadMetricQueryResult{}
    for cursor.Next(repo.ctx) {
        stats := new(result.WorkloadMetricQueryResult)
        if err := cursor.Decode(stats); err != nil {
            return nil, err
        }
        results = append(results, stats)
    }

    return results, nil
}

func (repo *LogRepo) filtersForWorkloadServiceMetric(params *param.WorkloadMetricQueryParam) (b.D, error) {
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
    
    if len(params.MetricNames) == 0 {
        return nil, errors.New("workload service metric query metric names not specified")
    }
    targetFilters := b.A{}
    for _, metricName := range params.MetricNames {
        condition := b.D{}
        if params.MetricTarget != "" {
            condition = append(condition, b.E{"data.metric.tags.metric_target", params.MetricTarget})
        }
        if metricName != "" {
            condition = append(condition, b.E{"data.metric.tags.metric_name", metricName})
        }
        if len(condition) > 0 {
            targetFilters = append(targetFilters, condition)
        }
    }
    filters = append(filters, b.E{"$or", targetFilters})

    for key, val := range params.Filters.Metadata {
        if val != "" {
            filters = append(filters, b.E{"data.metric.snapshots.metadata."+key, val})
        }
    }

    return filters, nil
}