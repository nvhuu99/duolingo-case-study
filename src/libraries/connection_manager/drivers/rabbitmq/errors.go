package rabbitmq

import "errors"

var (
	ErrConnectionType        = errors.New("provided with a connection not ampq.Channel")
	ErrConnectionArgsType    = errors.New("provided with an argument type that is not RabbitMQConnectionArgs")
	ErrInvalidConnectionArgs = errors.New("provided with invalid connection arguments")
	ErrConnectionIsClosed    = errors.New("connection is closed")
)
