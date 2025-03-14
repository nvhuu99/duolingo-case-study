package workdistributor

import "time"

type DistributorOptions struct {
	LockTimeOut			time.Duration
	DistributionSize	int
}

func DefaultDistributorOptions() *DistributorOptions {
	return &DistributorOptions{
		LockTimeOut: 10 * time.Second,
		DistributionSize: 100,
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