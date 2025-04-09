package messagequeue

type Dispatcher interface {
	// return the routing key for the publisher
	Dispatch(message string) string
}

/* Direct */

type Direct struct {
	Pattern string
}

func (d *Direct) Dispatch(message string) string {
	return d.Pattern
}

/* Fanout */

type FanOut struct {
}

func (d *FanOut) Dispatch(message string) string {
	return ""
}

/* Balancing */

type Balancing struct {
	Patterns []string
}

func (d *Balancing) Dispatch(message string) string {
	return ""
}
