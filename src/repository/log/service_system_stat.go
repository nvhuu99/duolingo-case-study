package log

import (
	ds "duolingo/repository/log/downsampling"
	"time"
)

type CpuUtil struct {
    StartLatencyMs int `json:"start_latency_ms" bson:"start_latency_ms"` 
    Value float64 `json:"value" bson:"value"` 
}

type SystemStats struct {
    CpuUtil []*CpuUtil `json:"cpu_util" bson:"cpu_util"`
}

type ServiceSystemStatReport struct {
    ServiceName      string          `json:"service_name" bson:"service_name"`

    Snapshots *SystemStats `json:"snapshots" bson:"snapshots"`

    MovingMedian *SystemStats `json:"moving_median" bson:"moving_median"`

    LTTB *SystemStats `json:"lttb" bson:"lttb"`

    Percentiles map[string]*SystemStats `json:"percentiles" bson:"percentiles"`
}

func (sts *ServiceSystemStatReport) Downsampling(workloadStart time.Time, reductionStep int) error {
    return sts.DownsamplingCPUUtil(workloadStart, reductionStep)
}

func (sts *ServiceSystemStatReport) DownsamplingCPUUtil(workloadStart time.Time, reductionStep int) error {
    datapoints := []*ds.DataPoint{}
    for _, snapshot := range sts.Snapshots.CpuUtil {
        datapoints = append(datapoints, ds.NewDataPoint(
            workloadStart.Add(time.Duration(snapshot.StartLatencyMs) * time.Millisecond), 
            snapshot.Value,
        ))
    }
    downsampling := new(ds.Downsampling).
                        WithStartTime(workloadStart).
                        WithReductionStep(int64(reductionStep)).
                        WithDatapoints(datapoints)

    fromDataPoints := func(datapoints []*ds.DataPoint) []*CpuUtil {
        result := []*CpuUtil{}
        for _, dp := range datapoints {
            latency := int(dp.GetTimestamp().Sub(workloadStart).Milliseconds())
            result = append(result, &CpuUtil{ latency, dp.GetValue() })
        }
        return result
    }
    
    movingAvg, err := downsampling.WithStrategy(new(ds.MovingAverage)).Result()
    if err != nil {
        return err
    }
    sts.MovingMedian = &SystemStats{ CpuUtil: fromDataPoints(movingAvg) }

    lttb, err := downsampling.WithStrategy(new(ds.LTTB)).Result()
    if err != nil {
        return err
    }
    sts.LTTB = &SystemStats{ CpuUtil: fromDataPoints(lttb) }

    p5, err5 := downsampling.WithStrategy(ds.NewPercentileStrategy(5)).Result()
    if err5 != nil {
        return err5
    }
    p25, err25 := downsampling.WithStrategy(ds.NewPercentileStrategy(25)).Result()
    if err25 != nil {
        return err25
    }
    p75, err75 := downsampling.WithStrategy(ds.NewPercentileStrategy(75)).Result()
    if err75 != nil {
        return err75
    }
    p95, err95 := downsampling.WithStrategy(ds.NewPercentileStrategy(95)).Result()
    if err95 != nil {
        return err95
    }
    sts.Percentiles = make(map[string]*SystemStats)
    sts.Percentiles["5"] = &SystemStats{ CpuUtil: fromDataPoints(p5) }
    sts.Percentiles["25"] = &SystemStats{ CpuUtil: fromDataPoints(p25) }
    sts.Percentiles["75"] = &SystemStats{ CpuUtil: fromDataPoints(p75) }
    sts.Percentiles["95"] = &SystemStats{ CpuUtil: fromDataPoints(p95) }

    return nil
}
