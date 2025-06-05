package pub_sub

type Topic struct {
	name string
}

func NewTopic(name string) *Topic {
	return &Topic{
		name: name,
	}
}

func (topic *Topic) GetName() string {
	return topic.name
}

func (topic *Topic) Equal(comparedTo *Topic) bool {
	return topic.GetName() == comparedTo.GetName()
}
