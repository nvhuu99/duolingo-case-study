package message

import (
	msg "duolingo/libraries/push_notification/message"
	"fmt"
	"time"

	fcm "firebase.google.com/go/v4/messaging"
)

type IOSMessage struct {
	*msg.Message
}

func NewIOSMessage(msg *msg.Message) *IOSMessage {
	return &IOSMessage{msg}
}

func (ios *IOSMessage) Build() *fcm.APNSConfig {
	return &fcm.APNSConfig{
		Headers: map[string]string{
			"apns-priority":   ios.getPriority(),
			"apns-expiration": ios.getExpiration(),
		},
		Payload: &fcm.APNSPayload{
			Aps: &fcm.Aps{
				Alert: &fcm.ApsAlert{
					Title: ios.Title,
					Body:  ios.Body,
				},
				Sound: ios.Sound,
			},
		},
	}
}

func (ios *IOSMessage) getExpiration() string {
	if ios.Expiration == time.Duration(0) {
		return ""
	}
	return fmt.Sprint(time.Now().Add(ios.Expiration).Unix())
}

func (ios *IOSMessage) getPriority() string {
	priorities := map[msg.Priority]string{
		msg.PriorityHigh:   "10",
		msg.PriorityNormal: "5",
	}
	if priority, found := priorities[ios.Priority]; found {
		return priority
	}
	return "5"
}
