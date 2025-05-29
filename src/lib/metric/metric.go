package metric

import (
	"context"
	"errors"
	"sort"
	"sync"
	"time"
)

type Metric struct {
	datapointsChan    chan *DataPoint
	snapshotTick      time.Duration
	datapointInterval time.Duration
	captureStatus     CaptureStatus

	collectors []Collector

	parentCtx context.Context
	ctx       context.Context
	cancel    context.CancelFunc

	mu sync.Mutex
}

func NewMetric(ctx context.Context, interval time.Duration, tick time.Duration) *Metric {
	return &Metric{
		parentCtx:         ctx,
		snapshotTick:      tick,
		datapointInterval: interval,
	}
}

func (m *Metric) AddCollector(c Collector) *Metric {
	m.collectors = append(m.collectors, c)
	return m
}

func (m *Metric) CaptureStart() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.captureStatus == CaptureStatusStarted {
		return errors.New(ErrMessages[ERR_CAPTURE_STARTED_ALREADY])
	}

	m.ctx, m.cancel = context.WithCancel(m.parentCtx)
	m.captureStatus = CaptureStatusStarted

	go m.capturing()

	return nil
}

func (m *Metric) CaptureEnd() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.captureStatus != CaptureStatusStarted {
		return errors.New(ErrMessages[ERR_CAPTURE_HAS_NOT_STARTED])
	}
	if m.captureStatus == CaptureStatusEnded {
		return nil
	}

	m.cancel()
	m.captureStatus = CaptureStatusEnded

	return nil
}

func (m *Metric) DataPointChannel() (<-chan *DataPoint, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.captureStatus != CaptureStatusStarted && m.captureStatus != CaptureStatusEnded {
		return nil, errors.New(ErrMessages[ERR_CAPTURE_HAS_NOT_STARTED])
	}

	return m.datapointsChan, nil
}

func (m *Metric) capturing() {
	m.datapointsChan = make(chan *DataPoint, 1000)
	defer close(m.datapointsChan)

	ticker := time.NewTicker(m.snapshotTick)
	defer ticker.Stop()

	timer := time.NewTimer(m.datapointInterval)
	defer timer.Stop()

	capture := func() {
		for _, collector := range m.collectors {
			collector.Capture()
		}
	}
	buffer := func() {
		for _, collector := range m.collectors {
			rawDatapoints := collector.Collect()
			for _, raw := range rawDatapoints {
				if len(raw.Snapshots) == 0 {
					continue
				}
				sort.Slice(raw.Snapshots, func(i, j int) bool {
					return raw.Snapshots[i].Timestamp.Before(raw.Snapshots[j].Timestamp)
				})
				end := raw.Snapshots[len(raw.Snapshots) - 1].Timestamp
				duration := m.snapshotTick * time.Duration(len(raw.Snapshots) - 1)
				dp := &DataPoint{
					EndTime: end,
					StartTime: end.Add(-duration),
					DurationMs: duration.Milliseconds(),
					IncrMs: m.snapshotTick.Milliseconds(),
					Count: len(raw.Snapshots),
					Snapshots: raw.Snapshots,
					Tags: raw.Tags,
				}
				m.datapointsChan <- dp
			}
		}
	}

	for {
		select {
		case <-m.ctx.Done():
			buffer()
			return
		case <-timer.C:
			buffer()
			timer.Reset(m.datapointInterval)
		case <-ticker.C:
			capture()
		}
	}
}
