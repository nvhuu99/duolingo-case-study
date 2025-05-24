package metric

import (
	"duolingo/lib/metric"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/mem"
)

type SystemStatsCollector struct {
	serviceName string
	snapshots map[string][]*metric.Snapshot
}

func NewSystemStatsCollector(serviceName string) *SystemStatsCollector {
	return &SystemStatsCollector{
		serviceName: serviceName,
		snapshots:  make(map[string][]*metric.Snapshot),
	}
}

func (c *SystemStatsCollector) Capture() {
	if cpuPercents, _ := cpu.Percent(0, false); len(cpuPercents) > 0  {
		c.snapshots["cpu_util"] = append(c.snapshots["cpu_util"], metric.NewSnapshot(cpuPercents[0]))
	}
	if vmem, err := mem.VirtualMemory(); err == nil {
		c.snapshots["memory_used_pct"] = append(c.snapshots["memory_used_pct"], metric.NewSnapshot(vmem.UsedPercent))
		c.snapshots["memory_used_mb"] = append(c.snapshots["memory_used_mb"], metric.NewSnapshot(float64(vmem.Used)/1024/1024))
	}
	if diskIO, err := disk.IOCounters(); err == nil {
		for name, partition := range diskIO {
			c.snapshots["disk_io"] = append(c.snapshots["disk_io"], metric.NewSnapshot(float64(partition.IoTime), "partition", name))
		}
	}
}

func (c *SystemStatsCollector) Collect() []*metric.DataPoint {
	defer func() { 
		c.snapshots = make(map[string][]*metric.Snapshot)
	}()
	return []*metric.DataPoint{
		metric.RawDataPoint(c.snapshots["cpu_util"], "metric_target", c.serviceName, "metric_name", "cpu_util"),
		metric.RawDataPoint(c.snapshots["memory_used_pct"], "metric_target", c.serviceName, "metric_name", "memory_used_pct"),
		metric.RawDataPoint(c.snapshots["memory_used_mb"], "metric_target", c.serviceName, "metric_name", "memory_used_mb"),
		metric.RawDataPoint(c.snapshots["disk_io"], "metric_target", c.serviceName, "metric_name", "disk_io"),
	}
}
