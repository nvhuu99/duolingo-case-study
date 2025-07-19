package firebase

import (
	"context"
	push_noti "duolingo/libraries/push_notification"
	"duolingo/libraries/push_notification/message"

	"firebase.google.com/go/v4"
	"google.golang.org/api/option"

	driver "duolingo/libraries/push_notification/drivers/firebase/message"
)

type FirebasePushNotiFactory struct {
	ctx         context.Context
	firebaseApp *firebase.App
}

func NewFirebasePushNotiFactory(ctx context.Context, credJson string) (*FirebasePushNotiFactory, error) {
	app, err := firebase.NewApp(ctx, nil,
		option.WithCredentialsJSON([]byte(credJson)),
	)
	if err != nil {
		return nil, err
	}
	factory := &FirebasePushNotiFactory{
		ctx:         ctx,
		firebaseApp: app,
	}
	return factory, nil
}

func (factory *FirebasePushNotiFactory) CreatePushService() (push_noti.PushService, error) {
	client, err := factory.firebaseApp.Messaging(factory.ctx)
	if err != nil {
		return nil, err
	}
	service := NewFirebasePushService(
		client,
		factory.CreateMessageBuilder(),
	)
	return service, nil
}

func (factory *FirebasePushNotiFactory) CreateMessageBuilder() message.MessageBuilder {
	return driver.NewFirebaseMessagebuilder()
}
