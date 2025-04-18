package event

import "sync"

type Publisher interface {
	Subscribe(string, Subcriber)
	SubscribeRegex(string, Subcriber)
	UnSubscribe(string, Subcriber)
	Notify(*sync.WaitGroup, string, any)
}
