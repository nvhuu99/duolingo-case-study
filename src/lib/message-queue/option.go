package messagequeue

import "time"

type ManagerOptions struct {	
	GraceTimeOut		time.Duration
	ConnectionTimeOut	time.Duration
	HearBeat			time.Duration
	KeepAlive			bool
}

type TopologyOptions struct {
	GraceTimeOut		time.Duration
	DeclareTimeOut		time.Duration
}

type PublisherOptions struct {
	Topic		string
	Dispatcher	Dispatcher

	GraceTimeOut		time.Duration
	WriteTimeOut		time.Duration
}

type ConsumerOptions struct {
	Queue				string
	GraceTimeOut		time.Duration
}

/* Manager */

func DefaultManagerOptions() *ManagerOptions {
	return &ManagerOptions {
		GraceTimeOut:		300 * time.Millisecond,
		ConnectionTimeOut:	60 * time.Second,
		HearBeat:			10 * time.Second,
		KeepAlive: 			true,
	}
}

/* Topology */

func DefaultTopologyOptions() *TopologyOptions {
	return &TopologyOptions {
		GraceTimeOut:		300 * time.Millisecond,
		DeclareTimeOut:		60 * time.Second,
	}
}

/* Consumer */

func DefaultConsumerOptions() *ConsumerOptions {
	return &ConsumerOptions {
		GraceTimeOut:		300 * time.Millisecond,
	}
}

func (opt *ConsumerOptions) WithQueue(queue string) *ConsumerOptions {
	opt.Queue = queue
	return opt
}

/* Publisher */

func DefaultPublisherOptions() *PublisherOptions {
	return &PublisherOptions {
		GraceTimeOut:		300 * time.Millisecond,
		WriteTimeOut:		5 * time.Second,
	}
}

func (opt *PublisherOptions) WithTopic(topic string) *PublisherOptions {
	opt.Topic = topic
	return opt
}

func (opt *PublisherOptions) WithFanOutDispatch() *PublisherOptions {
	opt.Dispatcher = &FanOut{}
	return opt
}

func (opt *PublisherOptions) WithBalancingDispatch(patterns... string) *PublisherOptions {
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
