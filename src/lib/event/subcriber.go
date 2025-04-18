package event

import "sync"

type Subcriber interface {
	SubcriberId() string
	Notified(*sync.WaitGroup, string, any)
}
