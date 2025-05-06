package metric

import (
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/mem"
)

type CPUStats struct {
	Util          float32 `json:"util"`
	IOTimeSeconds float64 `json:"io_time_seconds"`
}

type MemoryStats struct {
	UsedPct float32 `json:"used_pct"`
	FreePct float32 `json:"free_pct"`
	UsedMB  uint32  `json:"used_mb"`
	FreeMB  uint32  `json:"free_mb"`
}

type DiskIOStats struct {
	Device   string `json:"device"`
	IOTimeMs uint64 `json:"io_time_ms"`
}

type SystemStats struct {
	CPU    *CPUStats               `json:"cpu"`
	Memory *MemoryStats            `json:"memory"`
	DiskIO map[string]*DiskIOStats `json:"disk_io"`
}

type SystemStatsCollector struct {
}

func (c *SystemStatsCollector) Capture() any {
	systemStats := new(SystemStats)

	cpuPercents, _ := cpu.Percent(0, false)
	cpuTimes, _ := cpu.Times(false)
	if len(cpuPercents) > 0 && len(cpuTimes) > 0 {
		systemStats.CPU = new(CPUStats)
		systemStats.CPU.Util = float32(cpuPercents[0])
		systemStats.CPU.IOTimeSeconds = cpuTimes[0].Iowait
	}

	vmem, err := mem.VirtualMemory()
	if err == nil {
		systemStats.Memory = new(MemoryStats)
		systemStats.Memory.UsedPct = float32(vmem.UsedPercent)
		systemStats.Memory.FreePct = float32(100.0 - vmem.UsedPercent)
		systemStats.Memory.UsedMB = uint32(vmem.Used / 1024 / 1024)
		systemStats.Memory.FreeMB = uint32(vmem.Available / 1024 / 1024)
	}

	diskIO, err := disk.IOCounters()
	if err == nil {
		systemStats.DiskIO = make(map[string]*DiskIOStats)
		for name, partition := range diskIO {
			systemStats.DiskIO[name] = &DiskIOStats{
				Device:   name,
				IOTimeMs: partition.IoTime,
			}
		}

	}

	return systemStats
}
