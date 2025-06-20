package message

import (
	msg "duolingo/libraries/push_notification/message"
	fcm "firebase.google.com/go/v4/messaging"
	"time"
)

type AndroidMessage struct {
	*msg.Message
}

func NewAndroidMessage(msg *msg.Message) *AndroidMessage {
	return &AndroidMessage{msg}
}

func (android *AndroidMessage) Build() *fcm.AndroidConfig {
	return &fcm.AndroidConfig{
		Priority:    android.getPriority(),
		TTL:         android.getExpiration(),
		CollapseKey: android.CollapseKey,
		Notification: &fcm.AndroidNotification{
			Icon:       android.Icon,
			Sound:      android.Sound,
			Visibility: android.getVisibility(),
		},
	}
}

func (android *AndroidMessage) getExpiration() *time.Duration {
	if android.Expiration == time.Duration(0) {
		return nil
	}
	return &android.Expiration
}

func (android *AndroidMessage) getPriority() string {
	priorities := map[msg.Priority]string{
		msg.PriorityHigh:   "high",
		msg.PriorityNormal: "normal",
	}
	if priority, found := priorities[android.Priority]; found {
		return priority
	}
	return "normal"
}

func (android *AndroidMessage) getVisibility() fcm.AndroidNotificationVisibility {
	visibilities := map[msg.Visibility]any{
		msg.VisibilityPrivate: fcm.VisibilityPrivate,
		msg.VisibilityPublic:  fcm.VisibilityPublic,
		msg.VisibilitySecret:  fcm.VisibilitySecret,
	}
	visibility, found := visibilities[android.Visibility].(fcm.AndroidNotificationVisibility)
	if found {
		return visibility
	}
	return fcm.VisibilityPublic
}
