package metric

import (
	"time"
)

type Datapoint struct {
	Timestamp  time.Time `json:"timestamp"`
	Duration   uint64    `json:"duration_ms"`
	Count      uint16    `json:"count"`
	Sum        *Metric   `json:"sum"`
	Mean       *Metric   `json:"mean"`
	Percentile *Metric   `json:"percentile"`
	UpperBound *Metric   `json:"upperbound"`
	LowerBound *Metric   `json:"lowerbound"`
}

func NewDataPointFromMetrics(timestamp time.Time, duration time.Duration, metrics []*Metric) *Datapoint {
	if len(metrics) == 0 {
		return nil
	}

	dp := &Datapoint{
		Timestamp:  timestamp,
		Duration:   uint64(duration.Milliseconds()),
		Count:      uint16(len(metrics)),
		Sum:        NewMetric(),
		Mean:       NewMetric(),
		Percentile: NewMetric(),
		UpperBound: NewMetric(),
		LowerBound: NewMetric(),
	}

	cpuUtils := make([]float32, len(metrics))
	cpuIOTimes := make([]float64, len(metrics))
	memUsedPct := make([]float32, len(metrics))
	memFreePct := make([]float32, len(metrics))
	memUsedMB := make([]uint32, len(metrics))
	memFreeMB := make([]uint32, len(metrics))
	diskUtils := make(map[string][]float32)
	diskIOTimes := make(map[string][]uint64)
	diskDevices := []string{}
	for _, metric := range metrics {
		if metric.DiskIOMetrics != nil {
			for dev := range metric.DiskIOMetrics {
				diskDevices = append(diskDevices, dev)
				diskUtils[dev] = make([]float32, len(metrics))
				diskIOTimes[dev] = make([]uint64, len(metrics))
			}
		}
	}

	metricInterval := int64(duration.Milliseconds()) / int64(len(metrics))
	for i, m := range metrics {
		if m.CPUMetric != nil {
			cpuUtils[i] = m.CPUMetric.Util
			cpuIOTimes[i] = m.CPUMetric.IOTimeSeconds
		}
		if m.MemoryMetric != nil {
			memUsedPct[i] = m.MemoryMetric.UsedPct
			memFreePct[i] = m.MemoryMetric.FreePct
			memUsedMB[i] = m.MemoryMetric.UsedMB
			memFreeMB[i] = m.MemoryMetric.FreeMB
		}
		if m.DiskIOMetrics != nil {
			for dev, io := range m.DiskIOMetrics {
				diskIOTimes[dev][i] = io.IOTimeMs
				if i > 0 && metricInterval > 0 {
					delta := diskIOTimes[dev][i] - diskIOTimes[dev][i-1]
					diskUtils[dev][i] = float32(delta) / float32(metricInterval) * 100
				}
			}
		}
	}

	if len(cpuUtils) > 0 {
		dp.Sum.CPUMetric = &CPUMetric{
			Util:          float32(sum(cpuUtils)),
			IOTimeSeconds: float64(sum(cpuIOTimes)),
		}
		dp.Mean.CPUMetric = &CPUMetric{
			Util:          float32(mean(cpuUtils)),
			IOTimeSeconds: float64(mean(cpuIOTimes)),
		}
		dp.Percentile.CPUMetric = &CPUMetric{
			Util:          float32(percentile(cpuUtils, 0.9)),
			IOTimeSeconds: float64(percentile(cpuIOTimes, 0.9)),
		}
		dp.UpperBound.CPUMetric = &CPUMetric{
			Util:          float32(max(cpuUtils)),
			IOTimeSeconds: float64(max(cpuIOTimes)),
		}
		dp.LowerBound.CPUMetric = &CPUMetric{
			Util:          float32(min(cpuUtils)),
			IOTimeSeconds: float64(min(cpuIOTimes)),
		}
	}

	if len(memUsedPct) > 0 {
		dp.Sum.MemoryMetric = &MemoryMetric{
			UsedPct: float32(sum(memUsedPct)),
			FreePct: float32(sum(memFreePct)),
			UsedMB:  uint32(sum(memUsedMB)),
			FreeMB:  uint32(sum(memFreeMB)),
		}
		dp.Mean.MemoryMetric = &MemoryMetric{
			UsedPct: float32(mean(memUsedPct)),
			FreePct: float32(mean(memFreePct)),
			UsedMB:  uint32(mean(memUsedMB)),
			FreeMB:  uint32(mean(memFreeMB)),
		}
		dp.Percentile.MemoryMetric = &MemoryMetric{
			UsedPct: float32(percentile(memUsedPct, 0.9)),
			FreePct: float32(percentile(memFreePct, 0.9)),
			UsedMB:  uint32(percentile(memUsedMB, 0.9)),
			FreeMB:  uint32(percentile(memFreeMB, 0.9)),
		}
		dp.UpperBound.MemoryMetric = &MemoryMetric{
			UsedPct: float32(max(memUsedPct)),
			FreePct: float32(max(memFreePct)),
			UsedMB:  uint32(max(memUsedMB)),
			FreeMB:  uint32(max(memFreeMB)),
		}
		dp.LowerBound.MemoryMetric = &MemoryMetric{
			UsedPct: float32(min(memUsedPct)),
			FreePct: float32(min(memFreePct)),
			UsedMB:  uint32(min(memUsedMB)),
			FreeMB:  uint32(min(memFreeMB)),
		}
	}

	if len(diskUtils) > 0 && len(diskIOTimes) > 0 {
		for _, dev := range diskDevices {
			dp.Sum.DiskIOMetrics[dev] = new(DiskIOMetric)
			dp.Sum.DiskIOMetrics[dev].Device = dev
			dp.Sum.DiskIOMetrics[dev].Util = float32(sum(diskUtils[dev]))
			dp.Sum.DiskIOMetrics[dev].IOTimeMs = uint64(sum(diskIOTimes[dev]))

			dp.Mean.DiskIOMetrics[dev] = new(DiskIOMetric)
			dp.Mean.DiskIOMetrics[dev].Device = dev
			dp.Mean.DiskIOMetrics[dev].Util = float32(mean(diskUtils[dev]))
			dp.Mean.DiskIOMetrics[dev].IOTimeMs = uint64(mean(diskIOTimes[dev]))

			dp.Percentile.DiskIOMetrics[dev] = new(DiskIOMetric)
			dp.Percentile.DiskIOMetrics[dev].Device = dev
			dp.Percentile.DiskIOMetrics[dev].Util = float32(percentile(diskUtils[dev], 0.9))
			dp.Percentile.DiskIOMetrics[dev].IOTimeMs = uint64(percentile(diskIOTimes[dev], 0.9))

			dp.UpperBound.DiskIOMetrics[dev] = new(DiskIOMetric)
			dp.UpperBound.DiskIOMetrics[dev].Device = dev
			dp.UpperBound.DiskIOMetrics[dev].Util = float32(max(diskUtils[dev]))
			dp.UpperBound.DiskIOMetrics[dev].IOTimeMs = uint64(max(diskIOTimes[dev]))

			dp.LowerBound.DiskIOMetrics[dev] = new(DiskIOMetric)
			dp.LowerBound.DiskIOMetrics[dev].Device = dev
			dp.LowerBound.DiskIOMetrics[dev].Util = float32(min(diskUtils[dev]))
			dp.LowerBound.DiskIOMetrics[dev].IOTimeMs = uint64(min(diskIOTimes[dev]))
		}
	}

	return dp
}
