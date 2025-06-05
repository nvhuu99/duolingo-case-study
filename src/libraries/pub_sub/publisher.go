package pub_sub

type Publisher interface {
	Notify(topic *Topic, data any) error
}
