package main

import (
	"context"
	cnst "duolingo/constants"
	ps "duolingo/libraries/pub_sub"
	"duolingo/libraries/push_notification"
	container "duolingo/libraries/service_container"
	"duolingo/services/push_sender/bootstrap"
	"duolingo/services/push_sender/server"
	"time"
)

func main() {
	bootstrap.Bootstrap()

	notiSubscriber := container.MustResolveAlias[ps.Subscriber](cnst.PushNotiSubscriber)
	pushService := container.MustResolve[push_notification.PushService]()

	platforms := []string{
		"android",
		"ios",
	}
	bufferLimit := 500
	flushInterval := 100 * time.Millisecond
	sender := server.NewSender(
		notiSubscriber,
		pushService,
		platforms,
		bufferLimit,
		flushInterval,
	)

	err := sender.Start(context.Background())
	if err != nil {
		panic(err)
	}
}
