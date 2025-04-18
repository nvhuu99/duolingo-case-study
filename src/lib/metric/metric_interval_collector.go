package metric

import (
	"context"
	"errors"
	"time"
)

type MetricIntervalCollector struct {
	collector         *MetricCollector
	datapointInterval time.Duration
	snapshotTick      time.Duration
	ctx               context.Context
	cancel            context.CancelFunc
}

func NewMetricIntervalCollector(ctx context.Context, interval time.Duration, tick time.Duration) *MetricIntervalCollector {
	m := new(MetricIntervalCollector)
	m.ctx, m.cancel = context.WithCancel(ctx)
	m.datapointInterval = interval
	m.snapshotTick = tick
	return m
}

func (c *MetricIntervalCollector) StartInterval(flag CaptureFlag, callback func(*Datapoint)) {
	if c.collector != nil {
		c.collector.CaptureEnd()
	}
	c.collector = NewCollector(c.ctx, c.datapointInterval, c.snapshotTick)

	go c.capturing(flag, callback)
}

func (c *MetricIntervalCollector) StopInterval() error {
	if c.collector == nil || c.collector.captureStatus != CaptureStatusStarted {
		return errors.New(ErrMessages[ERR_CAPTURE_HAS_NOT_STARTED])
	}
	if c.collector.captureStatus == CaptureStatusEnded {
		return nil
	}

	c.cancel()

	return nil
}

func (c *MetricIntervalCollector) capturing(flag CaptureFlag, callback func(*Datapoint)) {
	if err := c.collector.CaptureStart(flag); err != nil {
		return
	}
	defer c.collector.CaptureEnd()

	dataPointChan, err := c.collector.DatapointChannel()
	if err != nil {
		return
	}

	for {
		select {
		case <-c.ctx.Done():
			return
		case datapoint := <-dataPointChan:
			go callback(datapoint)
		}
	}
}
