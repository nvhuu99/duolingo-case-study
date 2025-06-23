package bootstrap

import (
	"context"

	push_noti "duolingo/libraries/push_notification"
	driver "duolingo/libraries/push_notification/drivers/firebase"
	container "duolingo/libraries/service_container"
)

func BindPushService() {
	container.BindSingleton[push_noti.PushService](func(ctx context.Context) any {
		cred := `{
			"type": "service_account",
			"project_id": "your-project-id",
			"private_key_id": "your-private-key-id",
			"private_key": "-----BEGIN PRIVATE KEY-----\n<REDACTED>\n-----END PRIVATE KEY-----\n",
			"client_email": "your-service-account@your-project-id.iam.gserviceaccount.com",
			"client_id": "your-client-id",
			"auth_uri": "https://accounts.google.com/o/oauth2/auth",
			"token_uri": "https://oauth2.googleapis.com/token",
			"auth_provider_x509_cert_url": "https://www.googleapis.com/oauth2/v1/certs",
			"client_x509_cert_url": "https://www.googleapis.com/robot/v1/metadata/x509/your-service-account%40your-project-id.iam.gserviceaccount.com",
			"universe_domain": "googleapis.com"
		}`

		factory, factoryErr := driver.NewFirebasePushNotiFactory(ctx, cred)
		if factoryErr != nil {
			panic(factoryErr)
		}

		service, serviceErr := factory.CreatePushService()
		if serviceErr != nil {
			panic(serviceErr)
		}
		firebaseService, _ := service.(*driver.FirebasePushService)

		return firebaseService
	})
}
