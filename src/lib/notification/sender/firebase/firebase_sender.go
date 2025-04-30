package firebase

import (
	"context"
	"errors"

	"fmt"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"google.golang.org/api/option"

	noti "duolingo/lib/notification"
)

type FirebaseSender struct {
	client *messaging.Client
	ctx    context.Context
}

func NewFirebaseSender(ctx context.Context) *FirebaseSender {
	sender := new(FirebaseSender)
	sender.ctx = ctx
	return sender
}

func (sender *FirebaseSender) WithJsonCredentials(credentials string) error {
	app, err := firebase.NewApp(sender.ctx, nil,
		option.WithCredentialsJSON([]byte(credentials)),
	)
	if err != nil {
		return fmt.Errorf("%v - %w", noti.ErrMessages[noti.ERR_INVALID_CREDENTIALS], err)
	}

	client, err := app.Messaging(sender.ctx)
	if err != nil {
		return fmt.Errorf("%v - %w", noti.ErrMessages[noti.ERR_INVALID_CREDENTIALS], err)
	}

	sender.client = client

	return nil
}

func (sender *FirebaseSender) SendAll(title string, content string, deviceTokens []string) *noti.Result {
	if len(deviceTokens) == 0 {
		return &noti.Result{
			Success:       false,
			Error:         errors.New(noti.ErrMessages[noti.ERR_DEVICE_TOKENS_EMPTY]),
			FailureCount:  len(deviceTokens),
			SuccessCount:  0,
			FailureTokens: deviceTokens,
		}
	}

	crafted := &messaging.MulticastMessage{
		Notification: &messaging.Notification{
			Title: title,
			Body:  content,
		},
		Tokens: deviceTokens,
	}

	br, err := sender.client.SendEachForMulticast(sender.ctx, crafted)
	if err != nil {
		return &noti.Result{
			Success:       false,
			Error:         fmt.Errorf("%v - %w", noti.ErrMessages[noti.ERR_SEND_FAILURE], err),
			FailureCount:  len(deviceTokens),
			SuccessCount:  0,
			FailureTokens: deviceTokens,
		}
	}

	var failureCount int
	var failedTokens []string
	for idx, resp := range br.Responses {
		if !resp.Success {
			failedTokens = append(failedTokens, deviceTokens[idx])
			failureCount++
		}
	}

	return &noti.Result{
		Success:       true,
		SuccessCount:  len(deviceTokens) - failureCount,
		FailureCount:  failureCount,
		FailureTokens: failedTokens,
	}
}

func (sender *FirebaseSender) GetTokenLimit() int {
	const fireBaseTokenLimit = 500
	return fireBaseTokenLimit
}
