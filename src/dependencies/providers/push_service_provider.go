package providers

import (
	"context"
	"fmt"

	"duolingo/apps/push_sender/server/test/fakes"
	"duolingo/libraries/config_reader"
	container "duolingo/libraries/dependencies_container"
	push_noti "duolingo/libraries/push_notification"
	driver "duolingo/libraries/push_notification/drivers/firebase"
)

type PushServiceProvider struct {
}

func (provider *PushServiceProvider) Bootstrap(bootstrapCtx context.Context, scope string) {
	if scope == "test" {
		provider.registerFakePushServiceProvider()
	} else {
		provider.registerFirebasePushServiceProvider()
	}
}

func (provider *PushServiceProvider) Shutdown(shutdownCtx context.Context) {
}

func (provider *PushServiceProvider) registerFakePushServiceProvider() {
	container.BindSingleton[push_noti.PushService](func(ctx context.Context) any {
		return fakes.NewFakePushService()
	})
}

func (provider *PushServiceProvider) registerFirebasePushServiceProvider() {
	container.BindSingleton[push_noti.PushService](func(ctx context.Context) any {
		config := container.MustResolve[config_reader.ConfigReader]()
		cred := config.Get("firebase", "credentials")

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
