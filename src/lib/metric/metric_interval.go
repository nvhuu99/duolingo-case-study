package metric

// import (
// 	"context"
// 	"errors"
// 	"time"
// )

// type MetricInterval struct {
// 	collector         *Metric
// 	datapointInterval time.Duration
// 	snapshotTick      time.Duration
// 	ctx               context.Context
// 	cancel            context.CancelFunc
// }

// func NewMetricInterval(ctx context.Context, interval time.Duration, tick time.Duration) *MetricInterval {
// 	m := new(MetricInterval)
// 	m.ctx, m.cancel = context.WithCancel(ctx)
// 	m.datapointInterval = interval
// 	m.snapshotTick = tick
// 	return m
// }

// func (c *MetricInterval) StartInterval(callback func([]*DataPoint)) {
// 	if c.collector != nil {
// 		c.collector.CaptureEnd()
// 	}
// 	c.collector = NewMetric(c.ctx, c.datapointInterval, c.snapshotTick)

// 	go c.capturing(callback)
// }

// func (c *MetricInterval) StopInterval() error {
// 	if c.collector == nil || c.collector.captureStatus != CaptureStatusStarted {
// 		return errors.New(ErrMessages[ERR_CAPTURE_HAS_NOT_STARTED])
// 	}
// 	if c.collector.captureStatus == CaptureStatusEnded {
// 		return nil
// 	}

// 	c.cancel()

// 	return nil
// }

// func (c *MetricInterval) capturing(callback func([]*DataPoint)) {
// 	if err := c.collector.CaptureStart(); err != nil {
// 		return
// 	}
// 	defer c.collector.CaptureEnd()

// 	dataPointChan, err := c.collector.datapointChannel()
// 	if err != nil {
// 		return
// 	}

// 	for {
// 		select {
// 		case <-c.ctx.Done():
// 			return
// 		case datapoint := <-dataPointChan:
// 			go callback(datapoint)
// 		}
// 	}
// }
