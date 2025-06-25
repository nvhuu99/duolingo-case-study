package main

import (
	cnst "duolingo/constants"
	ps "duolingo/libraries/pub_sub"
	container "duolingo/libraries/service_container"
	"duolingo/services/message_input/bootstrap"
	"duolingo/services/message_input/server/handlers"
	"duolingo/services/message_input/server/requests"
)

func main() {
	bootstrap.Bootstrap()
	request, _ := requests.NewMessageInputRequest(
		"superbowl",
		"message title",
		"message body",
	)
	publisher := container.MustResolveAlias[ps.Publisher](cnst.MesgInputPublisher)
	handler := handlers.NewMessageInputRequestHandler(publisher)
	handler.Handle(request)
}
