package main

import (
	"context"
	"duolingo/constants"
	"duolingo/services/push_sender/bootstrap"
	"duolingo/services/push_sender/server"
	"time"
)

func main() {
	bootstrap.Bootstrap()

	topic := constants.TopicPushNotiMessages
	platforms := []string{
		"android", 
		"ios",
	}
	bufferLimit := 500
	flushInterval := 100 * time.Millisecond

	sender := server.NewSender(
		topic,
		platforms,
		bufferLimit,
		flushInterval,
	)

	err := sender.Start(context.Background())
	if err != nil {
		panic(err)
	}
}
