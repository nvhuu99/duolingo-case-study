package metric

import (
	"duolingo/lib/metric"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/mem"
)

type SystemStatsCollector struct {
	snapshots map[string][]*metric.Snapshot
}

func NewSystemStatsCollector() *SystemStatsCollector {
	return &SystemStatsCollector{
		snapshots:  make(map[string][]*metric.Snapshot),
	}
}

func (c *SystemStatsCollector) Capture() {
	if cpuPercents, _ := cpu.Percent(0, false); len(cpuPercents) > 0  {
		c.snapshots["cpu_util"] = append(c.snapshots["cpu_util"], metric.NewSnapshot(cpuPercents[0]))
	}
	if vmem, err := mem.VirtualMemory(); err == nil {
		c.snapshots["memory"] = append(c.snapshots["memory"], metric.NewSnapshot(vmem.UsedPercent, "is_used_percent"))
		// c.snapshots["memory"] = append(c.snapshots["memory"], metric.NewSnapshot(100.0-vmem.UsedPercent, "is_free_percent"))
		c.snapshots["memory"] = append(c.snapshots["memory"], metric.NewSnapshot(float64(vmem.Used)/1024/1024, "is_used_mb"))
		// c.snapshots["memory"] = append(c.snapshots["memory"], metric.NewSnapshot(float64(vmem.Available)/1024/1024, "is_free_mb"))
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
		metric.RawDataPoint(c.snapshots["cpu_util"], "target", "cpu_util"),
		metric.RawDataPoint(c.snapshots["memory"], "target", "memory"),
		metric.RawDataPoint(c.snapshots["disk_io"], "target", "disk_io"),
	}
}
