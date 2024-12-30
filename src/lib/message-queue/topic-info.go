package messagequeue

import (
	"errors"
)

type DistributeMethod string

const (
	QueueFanout   DistributeMethod = "queueFanout"   // a message will be
	QueueDispatch DistributeMethod = "queueDispatch" // messages are dispatched to each queue evenly

	errTopicEmpty    = "topic must be set before publishing messages"
	errQueueEmpty    = "queues list is empty"
	errMethodNotSet  = "distribute method must be set before publishing messages"
	errMethodUnknown = "unkown distribute method"
)

type TopicInfo struct {
	ConnectionString string
	Name             string
	Queues           []string
	Method           DistributeMethod

	queueIndex int
}

func (topic *TopicInfo) Next() (string, error) {
	if topic.Name == "" {
		return "", errors.New(errTopicEmpty)
	}
	if len(topic.Queues) == 0 {
		return "", errors.New(errQueueEmpty)
	}

	switch topic.Method {
	case QueueFanout:
		return "", nil
	case QueueDispatch:
		qName := topic.Queues[topic.queueIndex]
		if topic.queueIndex == len(topic.Queues)-1 {
			topic.queueIndex = 0
		} else {
			topic.queueIndex++
		}
		pattern, err := topic.Pattern(qName)
		if err != nil {
			return "", err
		}
		return pattern, nil
	}

	return "", errors.New(errMethodUnknown)
}

func (topic *TopicInfo) Pattern(qName string) (string, error) {
	switch topic.Method {
	case QueueDispatch:
		return qName, nil
	case QueueFanout:
		return "", nil
	}
	return "", errors.New(errMethodUnknown)
}
