package main

// import (
// 	"fmt"
// 	"path/filepath"
// 	"time"

// 	cnst "duolingo/constant"
// 	"duolingo/lib/log"
// 	lq "duolingo/lib/log/driver/log_query/local_file"
// 	"duolingo/lib/log/log_query"
// )

// func main() {
// 	logs:= logQuery().Filters(map[string]any{
// 		"context": map[string]any{
// 			"trace": map[string]any{
// 				"service_operation": cnst.BUILD_PUSH_NOTI_MESG,
// 			},
// 		},
// 	})

// 	sizes := []int{ 10000, 500, 100 }
// 	logGroup := map[int][]*log.Log

// 	err := logs.Each(func(item *log.Log) log_query.LoopAction {
// 		raw, _ := item.GetRaw("data.assignments")
// 		if asm, ok := raw.([]any); !ok || len(asm) == 0 {
// 			return log_query.LoopContinue
// 		}

// 		buildSize := item.GetInt("data.workload.num_of_units")

// 		for _, size := range sizes {
// 			if buildSize%size == 0 {

// 				break
// 			}
// 		}

// 		return log_query.LoopContinue
// 	})

// 	if err != nil {
// 		panic(err)
// 	}
// }

// func logQuery() *lq.LocalFileQuery {
// 	serviceDir, _ := filepath.Abs(filepath.Join(".", "service"))
// 	now := time.Now()
// 	from := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
// 	to := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 0, 0, now.Location())
// 	dir := filepath.Join(
// 		serviceDir,
// 		cnst.SV_NOTI_BUILDER,
// 		"storage", "log", "service",
// 		cnst.ServiceTypes[cnst.SV_NOTI_BUILDER],
// 	)
// 	return lq.FileQuery(dir, from, to)
// }
