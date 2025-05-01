package metric

import (
	"context"
	"errors"
	"sync"
	"time"
)

type Metric struct {
	datapointsChan    chan *DataPoint
	snapshotTick      time.Duration
	datapointInterval time.Duration
	captureStatus     CaptureStatus

	collectors map[string]Collector

	parentCtx context.Context
	ctx       context.Context
	cancel    context.CancelFunc

	mu sync.Mutex
}

func NewMetric(ctx context.Context, interval time.Duration, tick time.Duration) *Metric {
	collector := &Metric{
		parentCtx:         ctx,
		snapshotTick:      tick,
		datapointInterval: interval,
		collectors: make(map[string]Collector),
	}
	return collector
}

func (m *Metric) WithCollector(name string, c Collector) *Metric {	
	m.collectors[name] = c
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

func (m *Metric) Fetch() (*DataPoint, error) {
	dpChan, err := m.datapointChannel()
	if err != nil {
		return nil, err
	}

	to := time.After(m.datapointInterval + 200*time.Millisecond)
	select {
	case datapoint := <-dpChan:
		return datapoint, nil
	case <-to:
		return nil, errors.New(ErrMessages[ERR_NO_DATA_POINT_YET])
	}
}

func (m *Metric) datapointChannel() (<-chan *DataPoint, error) {
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

	start := time.Now()
	snapshots := make(map[string][]any)

	buffer := func() {
		if len(snapshots) == 0 {
			return
		}
		datapoint := &DataPoint{
			StartTime: start,
			EndTime: time.Now(),
			DurationMs: uint64(time.Since(start).Milliseconds()),
			IncrMs: uint64(m.snapshotTick.Milliseconds()),
			Count: uint16(len(snapshots)),
			Stats: snapshots,
		} 
		m.datapointsChan <- datapoint
		snapshots = make(map[string][]any)
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
			for name, collector := range m.collectors {
				snapshots[name] = append(snapshots[name], collector.Capture())
			}
		}
	}
}
