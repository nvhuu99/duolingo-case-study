package push_notification

import (
	"duolingo/libraries/push_notification/message"
)

type PushNotiFactory interface {
	CreatePushService() (PushService, error)
	CreateMessageBuilder() message.MessageBuilder
}
