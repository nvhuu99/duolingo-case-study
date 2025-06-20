package firebase

import (
	"context"
	"errors"

	"duolingo/libraries/push_notification/message"
	"duolingo/libraries/push_notification/results"

	fcm "firebase.google.com/go/v4/messaging"
)

var (
	ErrMessageDriverDriverMismatch = errors.New("message_builder.MessageBuilder driver mismatched")
)

type FirebasePushService struct {
	client  *fcm.Client
	builder message.MessageBuilder
}

func NewFirebasePushService(
	client *fcm.Client,
	builder message.MessageBuilder,
) *FirebasePushService {
	return &FirebasePushService{
		client:  client,
		builder: builder,
	}
}

func (service *FirebasePushService) SendMulticast(
	ctx context.Context,
	noti *message.Message,
	target *message.MulticastTarget,
) (
	*results.MulticastResult,
	error,
) {
	multicast, err := service.builder.BuildMulticast(noti, target)
	if err != nil {
		return nil, err
	}
	firebaseMulticast, ok := multicast.(*fcm.MulticastMessage)
	if !ok || firebaseMulticast == nil {
		panic(ErrMessageDriverDriverMismatch)
	}
	res, err := service.client.SendEachForMulticast(ctx, firebaseMulticast)
	if err != nil {
		return nil, err
	}
	return service.parseMulticastResponse(res, target), nil
}

func (service *FirebasePushService) parseMulticastResponse(
	res *fcm.BatchResponse,
	target *message.MulticastTarget,
) *results.MulticastResult {
	var failedTokens []string
	for i, resp := range res.Responses {
		if !resp.Success {
			failedTokens = append(failedTokens, target.DeviceTokens[i])
		}
	}
	return &results.MulticastResult{
		SuccessCount:  res.SuccessCount,
		FailureCount:  res.FailureCount,
		FailureTokens: failedTokens,
	}
}
