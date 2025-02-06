package rabbitmq

type Queue struct {
	name string
	bindings map[string]*Binding
}

func (queue *Queue) Bind(pattern string) *Binding {
	if _, found := queue.bindings[pattern]; !found {
		queue.bindings[pattern] = &Binding{ 
			Pattern: pattern,
		}
	}

	return queue.bindings[pattern]
}
