package metric

import (
	"context"
	"errors"
	"sync"
	"time"
)

type MetricCollector struct {
	datapointsChan    chan *Datapoint
	snapshotTick      time.Duration
	datapointInterval time.Duration
	captureStatus     CaptureStatus

	parentCtx context.Context
	ctx       context.Context
	cancel    context.CancelFunc

	mu sync.Mutex
}

func NewCollector(ctx context.Context, interval time.Duration, tick time.Duration) *MetricCollector {
	collector := &MetricCollector{
		parentCtx:         ctx,
		snapshotTick:      tick,
		datapointInterval: interval,
	}
	return collector
}

func (c *MetricCollector) CaptureStart(flag CaptureFlag) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.captureStatus == CaptureStatusStarted {
		return errors.New(ErrMessages[ERR_CAPTURE_STARTED_ALREADY])
	}

	c.ctx, c.cancel = context.WithCancel(c.parentCtx)
	c.captureStatus = CaptureStatusStarted

	go c.capturing(flag)

	return nil
}

func (c *MetricCollector) CaptureEnd() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.captureStatus != CaptureStatusStarted {
		return errors.New(ErrMessages[ERR_CAPTURE_HAS_NOT_STARTED])
	}
	if c.captureStatus == CaptureStatusEnded {
		return nil
	}

	c.cancel()
	c.captureStatus = CaptureStatusEnded

	return nil
}

func (c *MetricCollector) DatapointChannel() (<-chan *Datapoint, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.captureStatus != CaptureStatusStarted && c.captureStatus != CaptureStatusEnded {
		return nil, errors.New(ErrMessages[ERR_CAPTURE_HAS_NOT_STARTED])
	}

	return c.datapointsChan, nil
}

func (c *MetricCollector) Fetch() (*Datapoint, error) {
	dpChan, err := c.DatapointChannel()
	if err != nil {
		return nil, err
	}

	to := time.After(c.datapointInterval + 200*time.Millisecond)
	select {
	case datapoint := <-dpChan:
		return datapoint, nil
	case <-to:
		return nil, errors.New(ErrMessages[ERR_NO_DATA_POINT_YET])
	}
}

func (c *MetricCollector) capturing(flag CaptureFlag) {
	c.datapointsChan = make(chan *Datapoint, 1000)
	defer close(c.datapointsChan)

	ticker := time.NewTicker(c.snapshotTick)
	defer ticker.Stop()

	timer := time.NewTimer(c.datapointInterval)
	defer timer.Stop()

	start := time.Now()
	snapshots := []*Metric{}

	buffer := func() {
		if len(snapshots) == 0 {
			return
		}
		datapoint := NewDataPointFromMetrics(start, time.Since(start), snapshots)
		if datapoint == nil {
			return
		}
		c.datapointsChan <- datapoint
		snapshots = []*Metric{}
	}

	for {
		select {
		case <-c.ctx.Done():
			buffer()
			return
		case <-timer.C:
			buffer()
			timer.Reset(c.datapointInterval)
		case <-ticker.C:
			snapshots = append(snapshots, NewMetric().Capture(flag))
		}
	}
}
