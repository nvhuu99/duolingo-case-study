package messagequeue

import "time"

type ManagerOptions struct {
	GraceTimeOut      time.Duration
	ConnectionTimeOut time.Duration
	HearBeat          time.Duration
	KeepAlive         bool
}

type TopologyOptions struct {
	GraceTimeOut   time.Duration
	DeclareTimeOut time.Duration
	QueuesPurged   bool
}

type PublisherOptions struct {
	Topic      string
	Dispatcher Dispatcher

	GraceTimeOut time.Duration
	WriteTimeOut time.Duration
}

type ConsumerOptions struct {
	Queue        string
	GraceTimeOut time.Duration
}

/* Manager */

func DefaultManagerOptions() *ManagerOptions {
	return &ManagerOptions{
		GraceTimeOut:      300 * time.Millisecond,
		ConnectionTimeOut: 60 * time.Second,
		HearBeat:          10 * time.Second,
		KeepAlive:         true,
	}
}

func (opt *ManagerOptions) WithGraceTimeOut(duration time.Duration) *ManagerOptions {
	opt.GraceTimeOut = duration
	return opt
}

func (opt *ManagerOptions) WithConnectionTimeOut(duration time.Duration) *ManagerOptions {
	opt.ConnectionTimeOut = duration
	return opt
}

func (opt *ManagerOptions) WithHearBeat(duration time.Duration) *ManagerOptions {
	opt.HearBeat = duration
	return opt
}

func (opt *ManagerOptions) WithKeepAlive(flag bool) *ManagerOptions {
	opt.KeepAlive = flag
	return opt
}

/* Topology */

func DefaultTopologyOptions() *TopologyOptions {
	return &TopologyOptions{
		GraceTimeOut:   300 * time.Millisecond,
		DeclareTimeOut: 60 * time.Second,
		QueuesPurged:   false,
	}
}

func (opt *TopologyOptions) WithGraceTimeOut(duration time.Duration) *TopologyOptions {
	opt.GraceTimeOut = duration
	return opt
}

func (opt *TopologyOptions) WithDeclareTimeOut(duration time.Duration) *TopologyOptions {
	opt.DeclareTimeOut = duration
	return opt
}

func (opt *TopologyOptions) WithQueuesPurged(flag bool) *TopologyOptions {
	opt.QueuesPurged = flag
	return opt
}

/* Consumer */

func DefaultConsumerOptions() *ConsumerOptions {
	return &ConsumerOptions{
		GraceTimeOut: 300 * time.Millisecond,
	}
}

func (opt *ConsumerOptions) WithQueue(queue string) *ConsumerOptions {
	opt.Queue = queue
	return opt
}

func (opt *ConsumerOptions) WithGraceTimeOut(duration time.Duration) *ConsumerOptions {
	opt.GraceTimeOut = duration
	return opt
}

/* Publisher */

func DefaultPublisherOptions() *PublisherOptions {
	return &PublisherOptions{
		GraceTimeOut: 300 * time.Millisecond,
		WriteTimeOut: 5 * time.Second,
	}
}

func (opt *PublisherOptions) WithTopic(topic string) *PublisherOptions {
	opt.Topic = topic
	return opt
}

func (opt *PublisherOptions) WithWriteTimeOut(duration time.Duration) *PublisherOptions {
	opt.WriteTimeOut = duration
	return opt
}

func (opt *PublisherOptions) WithGraceTimeOut(duration time.Duration) *PublisherOptions {
	opt.GraceTimeOut = duration
	return opt
}

func (opt *PublisherOptions) WithFanOutDispatch() *PublisherOptions {
	opt.Dispatcher = &FanOut{}
	return opt
}

func (opt *PublisherOptions) WithBalancingDispatch(patterns ...string) *PublisherOptions {
	opt.Dispatcher = &Balancing{
		Patterns: patterns,
	}
	return opt
}

func (opt *PublisherOptions) WithDirectDispatch(pattern string) *PublisherOptions {
	opt.Dispatcher = &Direct{
		Pattern: pattern,
	}
	return opt
}
