package message

import (
	msg "duolingo/libraries/push_notification/message"
	"slices"

	fcm "firebase.google.com/go/v4/messaging"
)

type FirebaseMessagebuilder struct {
}

func NewFirebaseMessagebuilder() *FirebaseMessagebuilder {
	return &FirebaseMessagebuilder{}
}

func (builder *FirebaseMessagebuilder) BuildMulticast(
	message *msg.Message,
	target *msg.MulticastTarget,
) (any, error) {
	if err := message.Validate(); err != nil {
		return nil, err
	}
	if err := target.Validate(); err != nil {
		return nil, err
	}
	multicast := &fcm.MulticastMessage{
		Tokens: target.DeviceTokens,
		Notification: &fcm.Notification{
			Title: message.Title,
			Body:  message.Body,
		},
	}
	if slices.Contains(target.Platforms, msg.Android) {
		multicast.Android = NewAndroidMessage(message).Build()
	}
	if slices.Contains(target.Platforms, msg.IOS) {
		multicast.APNS = NewIOSMessage(message).Build()
	}
	return multicast, nil
}
