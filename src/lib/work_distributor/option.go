package work_distributor

import (
	"duolingo/lib/event"
	"time"
)

type DistributorOptions struct {
	LockTimeOut      time.Duration
	DistributionSize int
	Events           event.Publisher
}

func DefaultDistributorOptions() *DistributorOptions {
	return &DistributorOptions{
		LockTimeOut:      10 * time.Second,
		DistributionSize: 100,
		Events:           event.NewEventPublisher(),
	}
}

func (opts *DistributorOptions) WithLockTimeOut(duration time.Duration) *DistributorOptions {
	opts.LockTimeOut = duration
	return opts
}

func (opts *DistributorOptions) WithDistributionSize(size int) *DistributorOptions {
	opts.DistributionSize = size
	return opts
}

func (opts *DistributorOptions) WithEventPublisher(p event.Publisher) *DistributorOptions {
	opts.Events = p
	return opts
}
