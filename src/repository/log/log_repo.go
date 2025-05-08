package log

import (
	"context"
	"duolingo/lib/log"
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

// func (repo *LogRepo) AggregateWorkloadOptsExecTime(traceId string) ([]*WorkloadOptExcTimeAggr, error) {
//     workloadStart, err := repo.GetWorkloadStartTime(traceId)
//     if err != nil {
//         return nil, err
//     }

//     pipeline := mongo.Pipeline{
//         b.D{{"$match", b.D{
//             {"level", log.LevelDebug},
//             {"context.trace.trace_id", "dcb763f3-96fe-45e5-8154-de372e7448dd"},
//         }}},
//         b.D{{"$project", b.D{
//             {"service_name", "$context.trace.service_name"},
//             {"service_operation", "$context.trace.service_operation"},
//             {"start_time_latency_ms", b.D{
//                 {"$subtract", b.A{
//                     b.D{{"$toDate", "$data.metric.start_time"}},
//                     workloadStart,
//                 }},
//             }},
//             {"duration", "$data.metric.duration_ms"},
//         }}},
//         b.D{{"$group", b.D{
//             {"_id", b.D{
//                 {"service_name", "$service_name"},
//                 {"service_operation", "$service_operation"},
//             }},
//             {"start_time_latency_min", b.D{{"$min", "$start_time_latency_ms"}}},
//             {"durations", b.D{{"$push", "$duration"}}},
//             {"duration_min", b.D{{"$min", "$duration"}}},
//             {"duration_max", b.D{{"$max", "$duration"}}},
//             {"duration_avg", b.D{{"$avg", "$duration"}}},
//         }}},
//         b.D{{"$addFields", b.D{
//             {"durations_sorted", b.D{
//                 {"$sortArray", b.D{
//                     {"input", "$durations"},
//                     {"sortBy", 1},
//                 }},
//             }},
//         }}},
//         b.D{{"$addFields", b.D{
//             {"count", b.D{{"$size", "$durations_sorted"}}},
//             {"median", b.D{{"$let", b.D{
//                 {"vars", b.D{
//                     {"half", b.D{{"$divide", b.A{b.D{{"$size", "$durations_sorted"}}, 2}}}},
//                 }},
//                 {"in", b.D{
//                     {"$cond", b.A{
//                         b.D{{"$eq", b.A{b.D{{"$mod", b.A{"$$half", 1}}}, 0}}},
//                         b.D{{"$avg", b.A{
//                             b.D{{"$arrayElemAt", b.A{"$durations_sorted", b.D{{"$subtract", b.A{"$$half", 1}}}}}},
//                             b.D{{"$arrayElemAt", b.A{"$durations_sorted", "$$half"}}},
//                         }}},
//                         b.D{{"$arrayElemAt", b.A{"$durations_sorted", b.D{{"$floor", "$$half"}}}}}, // Added comma here
//                     }},
//                 }},
//             }}}},
//             {"percentile_5", b.D{{"$arrayElemAt", b.A{
//                 "$durations_sorted",
//                 b.D{{"$floor", b.D{{"$multiply", b.A{0.05, b.D{{"$subtract", b.A{b.D{{"$size", "$durations_sorted"}}, 1}}}}}}}},
//             }}}},
//             {"percentile_25", b.D{{"$arrayElemAt", b.A{
//                 "$durations_sorted",
//                 b.D{{"$floor", b.D{{"$multiply", b.A{0.25, b.D{{"$subtract", b.A{b.D{{"$size", "$durations_sorted"}}, 1}}}}}}}},
//             }}}},
//             {"percentile_50", b.D{{"$arrayElemAt", b.A{
//                 "$durations_sorted",
//                 b.D{{"$floor", b.D{{"$multiply", b.A{0.5, b.D{{"$subtract", b.A{b.D{{"$size", "$durations_sorted"}}, 1}}}}}}}},
//             }}}},
//             {"percentile_75", b.D{{"$arrayElemAt", b.A{
//                 "$durations_sorted",
//                 b.D{{"$floor", b.D{{"$multiply", b.A{0.75, b.D{{"$subtract", b.A{b.D{{"$size", "$durations_sorted"}}, 1}}}}}}}},
//             }}}},
//             {"percentile_95", b.D{{"$arrayElemAt", b.A{
//                 "$durations_sorted",
//                 b.D{{"$floor", b.D{{"$multiply", b.A{0.95, b.D{{"$subtract", b.A{b.D{{"$size", "$durations_sorted"}}, 1}}}}}}}},
//             }}}},
//         }}},
//         b.D{{"$project", b.D{
//             {"_id", 0},
//             {"service_name", "$_id.service_name"},
//             {"service_operation", "$_id.service_operation"},
//             {"start_time_latency", b.D{
//                 {"min", "$start_time_latency_min"},
//             }},
//             {"duration", b.D{
//                 {"count", "$count"},
//                 {"min", "$duration_min"},
//                 {"max", "$duration_max"},
//                 {"avg", "$duration_avg"},
//                 {"median", "$median"},
//                 {"percentile_5", "$percentile_5"},
//                 {"percentile_25", "$percentile_25"},
//                 {"percentile_50", "$percentile_50"},
//                 {"percentile_75", "$percentile_75"},
//                 {"percentile_95", "$percentile_95"},
//             }},
//         }}},
//     }
//     collection := repo.client.Database(repo.database).Collection("log")
//     cursor, err := collection.Aggregate(repo.ctx, pipeline)
//     if err != nil {
//         return nil, err
//     }
//     defer cursor.Close(repo.ctx)

//     result := []*WorkloadOptExcTimeAggr{}
//     for cursor.Next(repo.ctx) {
//         aggr := new(WorkloadOptExcTimeAggr)
//         if err := cursor.Decode(aggr); err != nil {
//             return nil, err
//         }
//         result = append(result, aggr)
//     }
    
//     return result, nil
// }

func (repo *LogRepo) ListWorkloadOperations(traceId string) ([]*WorkloadOperation, error) {
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

    result := []*WorkloadOperation{}
    for cursor.Next(repo.ctx) {
        aggr := new(WorkloadOperation)
        if err := cursor.Decode(aggr); err != nil {
            return nil, err
        }
        result = append(result, aggr)
    }
    
    return result, nil
}

func (repo *LogRepo) GetWorkloadOptsExecTimeSpans(traceId string) ([]*WorkloadOperationExecTimeSpan, error) {
    workloadStart, err := repo.GetWorkloadStartTime(traceId)
    if err != nil {
        return nil, err
    }
    pipeline := mongo.Pipeline{
        b.D{{"$match", b.D{
            {"level", log.LevelDebug},
            {"context.trace.trace_id", traceId},
        }}},
        b.D{{"$project", b.D{
            {"service_name", "$context.trace.service_name"},
            {"service_operation", "$context.trace.service_operation"},
            {"operation_start_latency_ms", b.D{
                {"$subtract", b.A{
                    b.D{{"$toDate", "$data.metric.start_time"}},
                    workloadStart,
                }},
            }},
            {"operation_end_latency_ms", b.D{
                {"$subtract", b.A{
                    b.D{{"$toDate", "$data.metric.end_time"}},
                    workloadStart,
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

    result := []*WorkloadOperationExecTimeSpan{}
    for cursor.Next(repo.ctx) {
        aggr := new(WorkloadOperationExecTimeSpan)
        if err := cursor.Decode(aggr); err != nil {
            return nil, err
        }
        result = append(result, aggr)
    }
    
    return result, nil
}

func (repo *LogRepo) GetWorkloadStartTime(traceId string) (*time.Time, error) {
    collection := repo.client.Database(repo.database).Collection("log")
    pipeline := mongo.Pipeline{
        {{"$match", b.D{
            {"level", log.LevelDebug},
            {"context.trace.trace_id", traceId},
        }}},
        {{"$group", b.D{
            {"_id", nil},
            {"min_start_time", b.D{{"$min", b.D{{"$toDate", "$data.metric.start_time"}}}}},
        }}},
    }
    cursor, err := collection.Aggregate(repo.ctx, pipeline)
    if err != nil {
        return nil, err
    }
    defer cursor.Close(repo.ctx)
    
    var result struct {
        MinStartTime time.Time `bson:"min_start_time"`
    }
    if cursor.Next(repo.ctx) {
        if err := cursor.Decode(&result); err != nil {
            return nil, err
        }
        return &result.MinStartTime, nil
    }
    return nil, fmt.Errorf("workload not found (trace_id: \"%v\")", traceId)
}
