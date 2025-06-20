package message

type MessageBuilder interface {
	BuildMulticast(msg *Message, target *MulticastTarget) (any, error)
}
