package server

import (
	"context"
	"duolingo/libraries/buffer"
	"duolingo/libraries/pub_sub"
	"duolingo/libraries/push_notification"
	"duolingo/libraries/push_notification/message"
	"duolingo/libraries/push_notification/results"
	"duolingo/libraries/service_container"
	"duolingo/models"
	"time"
)

type Sender struct {
	// Subscribe to the topic for receiving incoming push notification messages
	topic string
	pub_sub.Subscriber

	// Sending notifications to supported platforms (e.g., Android, IOS).
	platforms []string
	push_notification.PushService

	// The PushService may limit the number of device tokens that can be included
	// in each send request, or the system may need to manage the sending rate.
	// As a result, the Sender cannot submit a send request to the PushService for
	// every incoming push notification message received by the Subscriber. Instead,
	// each message is stored in a token buffer. When the buffer reaches its size
	// limit, the Sender is then able to flush the tokens and submit a send request
	// to the PushService.
	*buffer.BufferGroup[models.MessageInput, string]
}

func NewSender(
	topic string,
	platforms []string,
	bufferLimit int,
	bufferInterval time.Duration,
) *Sender {
	sub := service_container.MustResolve[pub_sub.Subscriber]()
	grp := buffer.NewBufferGroup[models.MessageInput, string]()
	grp.SetLimit(bufferLimit).SetInterval(bufferInterval)
	return &Sender{
		Subscriber:  sub,
		BufferGroup: grp,
	}
}

func (sender *Sender) Start(ctx context.Context) error {
	// When the buffer reaches size limit, flush the tokens and submit a send
	// request to the PushService.
	sender.SetConsumeFunc(false, func(input models.MessageInput, tokens []string) {
		sender.sendPushNoti(ctx, &input, tokens)
	})
	// Stored incoming push notification message in a token buffer
	return sender.Consuming(ctx, sender.topic, func(s string) pub_sub.ConsumeAction {
		sender.bufferMessage(models.PushNotiMessageDecode([]byte(s)))
		return pub_sub.ActionAccept
	})
}

func (sender *Sender) bufferMessage(message *models.PushNotiMessage) {
	sender.BufferGroup.Write(
		*message.MessageInput,
		message.GetTargetTokens(sender.platforms)...,
	)
}

func (sender *Sender) sendPushNoti(
	ctx context.Context,
	input *models.MessageInput,
	tokens []string,
) (
	*results.MulticastResult,
	error,
) {
	noti := &message.Message{
		Title: input.Title,
		Body:  input.Body,
	}
	target := &message.MulticastTarget{
		DeviceTokens: tokens,
		Platforms:    message.Platforms(sender.platforms...),
	}
	return sender.SendMulticast(ctx, noti, target)
}
