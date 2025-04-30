package event

type Publisher interface {
	Subscribe(string, Subcriber)
	SubscribeRegex(string, Subcriber)
	UnSubscribe(string, Subcriber)
	Notify(bool, string, any)
}
