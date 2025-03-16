package firebase

import (
	"context"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"google.golang.org/api/option"

	noti "duolingo/lib/notification"
)

type FirebaseSender struct {
	client *messaging.Client
	ctx context.Context
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
		return err
	}
	
	client, err := app.Messaging(sender.ctx)
	if err != nil {
		return err
	}

	sender.client = client

	return nil
}

func (sender *FirebaseSender) SendAll(message *noti.Message, deviceTokens []string) (*noti.Result, error) {
	crafted := &messaging.MulticastMessage{
		Notification: &messaging.Notification{
			Title: message.Title,
			Body: message.Body,
		},
		Tokens: deviceTokens,
	}
	
	br, err := sender.client.SendEachForMulticast(sender.ctx, crafted)
	if err != nil {
		return nil, err
	}

	var result noti.Result
	var failedTokens []string
	for idx, resp := range br.Responses {
		if !resp.Success {
			result.FailureTokens = append(failedTokens, deviceTokens[idx])
			result.FailureCount++
		}
	}

	return &result, nil
}
