package event_data

import (
	"duolingo/lib/notification"
	"duolingo/model"
)

type SendPushNotification struct {
	OptId    string
	PushNoti *model.PushNotiMessage
	Result   *notification.Result
}
