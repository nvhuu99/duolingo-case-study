package server

import (
	"context"
	"log"
	"sync"
	"time"

	"duolingo/libraries/buffer"
	"duolingo/libraries/config_reader"
	container "duolingo/libraries/dependencies_container"
	tq "duolingo/libraries/message_queue/task_queue"
	push_noti "duolingo/libraries/push_notification"
	"duolingo/libraries/push_notification/message"
	"duolingo/models"
)

type Sender struct {
	// Consumer receiving incoming push notification task
	pushNotiConsumer tq.TaskConsumer

	// Sending notifications to supported platforms (e.g., Android, IOS).
	platforms   []string
	pushService push_noti.PushService

	// The PushService may limit the number of device tokens that can be included
	// in each send request, or the system may need to manage the sending rate.
	// As a result, the Sender cannot submit a send request to the PushService for
	// every incoming push notification message received by the Subscriber. Instead,
	// each message is stored in a token buffer. When the buffer reaches its size
	// limit, the Sender is then able to flush the tokens and submit a send request
	// to the PushService.
	buffer *buffer.BufferGroup[models.MessageInput, string]

	// Sender operations are executed asynchronously, any errors occur might be
	// sent to this channel as a fallback handling.
	errChan chan error
}

func NewSender() *Sender {
	config := container.MustResolve[config_reader.ConfigReader]()
	platforms := config.GetArr("push_sender", "supported_platforms")
	bufferLimit := config.GetInt("push_sender", "buffer_limit")
	bufferInterval := time.Duration(config.GetInt("push_sender", "flush_duration")) * time.Millisecond
	grp := buffer.NewBufferGroup[models.MessageInput, string]()
	grp.SetLimit(bufferLimit).SetInterval(bufferInterval)

	pushNotiConsumer := container.MustResolveAlias[tq.TaskConsumer]("push_notifications_consumer")
	pushService := container.MustResolve[push_noti.PushService]()

	return &Sender{
		pushNotiConsumer: pushNotiConsumer,
		pushService:      pushService,
		buffer:           grp,
		platforms:        platforms,
		errChan:          make(chan error, 100),
	}
}

func (sender *Sender) Start(serverCtx context.Context) {
	ctx, cancel := context.WithCancel(serverCtx)
	defer cancel()

	wg := new(sync.WaitGroup)
	wg.Add(2)

	go sender.handleErrChannel(wg, ctx)

	go func() {
		defer cancel()
		defer wg.Done()
		// When the buffer reaches size limit, flush the tokens and submit a send
		// request to the PushService.
		sender.buffer.SetConsumeFunc(false, sender.sendPushNoti)
		// Stored incoming push notifications in a token buffer
		err := sender.pushNotiConsumer.Consuming(ctx, sender.bufferTokens)
		if err != nil {
			panic(err)
		}
	}()

	wg.Wait()
}

func (sender *Sender) handleErrChannel(wg *sync.WaitGroup, ctx context.Context) {
	defer wg.Done()
	for {
		select {
		case <-ctx.Done():
			return
		case err := <-sender.errChan:
			if err != nil {
				log.Println("push sender err", err)
			}
		}
	}
}

func (sender *Sender) bufferTokens(ctx context.Context, serialized string) {
	msg := models.PushNotiMessageDecode([]byte(serialized))
	if err := msg.Validate(); err != nil {
		sender.errChan <- err
		return
	}
	sender.buffer.DeclareGroup(ctx, *msg.MessageInput)
	sender.buffer.Write(*msg.MessageInput, msg.GetTargetTokens(sender.platforms)...)
	log.Println("token buffered:", serialized)
}

func (sender *Sender) sendPushNoti(
	ctx context.Context,
	input models.MessageInput,
	tokens []string,
) {
	noti := &message.Message{
		Title: input.Title,
		Body:  input.Body,
	}
	target := &message.MulticastTarget{
		DeviceTokens: tokens,
		Platforms:    message.Platforms(sender.platforms...),
	}
	if _, err := sender.pushService.SendMulticast(ctx, noti, target); err != nil {
		sender.errChan <- err
		return
	}
	log.Println("push notification sent:", string(input.Encode()))
}
