package rabbitmq

type Topic struct {
	name   string
	queues map[string]*Queue
}

func (topic *Topic) Queue(name string) *Queue {
	if _, found := topic.queues[name]; !found {
		topic.queues[name] = &Queue{
			name:     name,
			bindings: make(map[string]*Binding),
		}
	}

	return topic.queues[name]
}
