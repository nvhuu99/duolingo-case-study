package push_notification

import (
	"context"
	"duolingo/libraries/push_notification/message"
	"duolingo/libraries/push_notification/results"
)

type PushService interface {
	SendMulticast(ctx context.Context, noti *message.Message, target *message.MulticastTarget) (
		*results.MulticastResult,
		error,
	)
}
