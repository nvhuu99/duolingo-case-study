package workdistributor

import "time"

type DistributorOptions struct {
	LockTimeOut time.Duration
}

func DefaultDistributorOptions() *DistributorOptions {
	return &DistributorOptions{
		LockTimeOut: 10 * time.Second,
	}
}

func (opts *DistributorOptions) WithLockTimeOut(duration time.Duration) *DistributorOptions {
	opts.LockTimeOut = duration
	return opts
}