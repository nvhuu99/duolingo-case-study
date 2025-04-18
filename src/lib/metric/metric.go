package metric

import (
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/mem"
)

type CPUMetric struct {
	Util     float32 `json:"util"`
	IOTimeMs uint32  `json:"io_time_ms"`
}

type MemoryMetric struct {
	UsedPct float32 `json:"used_pct"`
	FreePct float32 `json:"free_pct"`
	UsedMB  uint32  `json:"used_mb"`
	FreeMB  uint32  `json:"free_mb"`
}

type DiskIOMetric struct {
	Device   string  `json:"device"`
	Util     float32 `json:"util"`
	IOTimeMs uint32  `json:"io_time_ms"`
}

type Metric struct {
	CPUMetric     *CPUMetric               `json:"cpu"`
	MemoryMetric  *MemoryMetric            `json:"memory"`
	DiskIOMetrics map[string]*DiskIOMetric `json:"disk_io"`
}

func NewMetric() *Metric {
	return &Metric{
		DiskIOMetrics: make(map[string]*DiskIOMetric),
	}
}

func (m *Metric) Capture(flag CaptureFlag) *Metric {
	if flag&CaptureCPU != 0 {
		m.CaptureCPU()
	}
	if flag&CaptureMemory != 0 {
		m.CaptureMemory()
	}
	if flag&CaptureDisksIO != 0 {
		m.CaptureDiskIO()
	}
	return m
}

func (m *Metric) CaptureCPU() *Metric {
	cpuPercents, _ := cpu.Percent(0, false)
	cpuTimes, _ := cpu.Times(false)
	if len(cpuPercents) > 0 && len(cpuTimes) > 0 {
		m.CPUMetric = new(CPUMetric)
		m.CPUMetric.Util = float32(cpuPercents[0])
		m.CPUMetric.IOTimeMs = uint32(cpuTimes[0].Iowait * 1000)
	}
	return m
}

func (m *Metric) CaptureMemory() *Metric {
	vmem, err := mem.VirtualMemory()
	if err == nil {
		m.MemoryMetric = new(MemoryMetric)
		m.MemoryMetric.UsedPct = float32(vmem.UsedPercent)
		m.MemoryMetric.FreePct = float32(100.0 - vmem.UsedPercent)
		m.MemoryMetric.UsedMB = uint32(vmem.Used / 1024 / 1024)
		m.MemoryMetric.FreeMB = uint32(vmem.Available / 1024 / 1024)
	}
	return m
}

func (m *Metric) CaptureDiskIO() *Metric {
	diskIO, err := disk.IOCounters()
	if err == nil {
		for name, partition := range diskIO {
			m.DiskIOMetrics[name] = &DiskIOMetric{
				Device:   name,
				IOTimeMs: uint32(partition.IoTime),
			}
		}
	}
	return m
}
