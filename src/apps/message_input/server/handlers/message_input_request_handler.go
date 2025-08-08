package handlers

import (
	container "duolingo/libraries/dependencies_container"
	ps "duolingo/libraries/message_queue/pub_sub"
	rest "duolingo/libraries/restful"
	"duolingo/models"
)

type MessageInputRequestHandler struct {
	inputPublisher ps.Publisher
}

func NewMessageInputRequestHandler() *MessageInputRequestHandler {
	publisher := container.MustResolveAlias[ps.Publisher]("message_input_publisher")
	return &MessageInputRequestHandler{
		inputPublisher: publisher,
	}
}

func (handler *MessageInputRequestHandler) Handle(req *rest.Request, res *rest.Response) {
	if valid, validations := handler.validate(req); !valid {
		res.BadRequest("invalid arguments", validations)
		return
	}
	message := models.NewMessageInput(
		req.PathArg("campaign").String(),
		req.Input("title").String(),
		req.Input("body").String(),
	)

	reqCtx := req.Context()
	err := handler.inputPublisher.NotifyMainTopic(reqCtx, string(message.Encode()))
	if err != nil {
		res.ServerErr("failed to input campaign message")
	} else {
		res.Ok("", message)
	}
}

func (handler *MessageInputRequestHandler) validate(req *rest.Request) (bool, map[string]string) {
	validations := make(map[string]string)
	if req.PathArg("campaign").String() == "" {
		validations["campaign"] = "campaign must not empty"
	}
	if req.Input("title").String() == "" {
		validations["title"] = "message title must not empty"
	}
	if req.Input("body").String() == "" {
		validations["body"] = "message body must not empty"
	}
	return len(validations) == 0, validations
}
