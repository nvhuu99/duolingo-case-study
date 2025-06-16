package main

import (
	"duolingo/services/message_input/bootstrap"
	"duolingo/services/message_input/server/handlers"
	"duolingo/services/message_input/server/requests"
)

func main() {
	bootstrap.Bootstrap()

	request, validationErr := requests.NewMessageInputRequest(
		"superbowl",
		"hello world",
		"nobody care",
	)
	if validationErr != nil {
		panic(validationErr)
	}

	err := handlers.HandleMessageInputRequest(request)
	if err != nil {
		panic(err)
	}
}
