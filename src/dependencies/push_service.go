package dependencies

import (
	"context"
	"fmt"

	"duolingo/apps/push_sender/server/test/fakes"
	"duolingo/libraries/config_reader"
	container "duolingo/libraries/dependencies_container"
	push_noti "duolingo/libraries/push_notification"
	driver "duolingo/libraries/push_notification/drivers/firebase"
)

type PushService struct {
	config config_reader.ConfigReader
}

func NewPushService() *PushService {
	return &PushService{}
}

func (provider *PushService) Shutdown() {
}

func (provider *PushService) Bootstrap(scope string) {
	provider.config = container.MustResolve[config_reader.ConfigReader]()

	if scope == "test" {
		provider.bindFakePushService()
		return
	}

	provider.bindFirebasePushService()
}

func (provider *PushService) bindFirebasePushService() {
	container.BindSingleton[push_noti.PushService](func(ctx context.Context) any {
		cred := provider.config.Get("firebase", "credentials")
		factory, factoryErr := driver.NewFirebasePushNotiFactory(ctx, cred)
		if factoryErr != nil {
			panic(fmt.Errorf("failed to setup push service with error: %v ", factoryErr))
		}

		service, serviceErr := factory.CreatePushService()
		if serviceErr != nil {
			panic(fmt.Errorf("failed to setup push service with error: %v ", serviceErr))
		}

		return service
	})
}

func (provider *PushService) bindFakePushService() {
	container.BindSingleton[push_noti.PushService](func(ctx context.Context) any {
		return fakes.NewFakePushService()
	})
}
