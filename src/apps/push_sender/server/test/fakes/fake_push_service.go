package fakes

import (
	"context"
	"duolingo/libraries/push_notification/message"
	"duolingo/libraries/push_notification/results"
)

type FakePushService struct {
	messageChan chan *FakeMessage
}

func NewFakePushService() *FakePushService {
	return &FakePushService{
		messageChan: make(chan *FakeMessage, 10),
	}
}

func (service *FakePushService) SendMulticast(
	ctx context.Context,
	noti *message.Message,
	target *message.MulticastTarget,
) (
	*results.MulticastResult,
	error,
) {
	platforms := []string{}
	for i := range target.Platforms {
		platforms = append(platforms, string(target.Platforms[i]))
	}
	fakeMessage := &FakeMessage{
		Title:     noti.Title,
		Body:      noti.Body,
		Tokens:    target.DeviceTokens,
		Platforms: platforms,
	}

	service.messageChan <- fakeMessage

	result := &results.MulticastResult{
		SuccessCount: len(target.DeviceTokens),
	}

	return result, nil
}

func (service *FakePushService) GetMesgChan() <-chan *FakeMessage {
	return service.messageChan
}
