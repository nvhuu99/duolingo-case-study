package rabbitmq

import (
	"context"
	"duolingo/libraries/connection_manager"
	"errors"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitMQConnectionProxy struct {
	ctx            context.Context
	connectionArgs *RabbitMQConnectionArgs
}

func NewRabbitMQConnectionProxy(ctx context.Context) *RabbitMQConnectionProxy {
	return &RabbitMQConnectionProxy{ctx: ctx}
}

/* Implement connection_manager.ConnectionProxy interface */

func (proxy *RabbitMQConnectionProxy) SetArgsPanicIfInvalid(args any) {
	rabbitMQArgs, ok := args.(*RabbitMQConnectionArgs)
	if !ok {
		panic(ErrConnectionArgsType)
	}
	if rabbitMQArgs.GetHost() == "" || rabbitMQArgs.GetPort() == "" {
		panic(ErrInvalidConnectionArgs)
	}
	if rabbitMQArgs.GetURI() == "" {
		address := fmt.Sprintf("%v:%v", rabbitMQArgs.GetHost(), rabbitMQArgs.GetPort())
		credentials := fmt.Sprintf("%v:%v", rabbitMQArgs.GetUser(), rabbitMQArgs.GetPassword())
		uri := fmt.Sprintf("amqp://%v/", address)
		if credentials != ":" {
			uri = fmt.Sprintf("amqp://%v@%v/", credentials, address)
		}
		rabbitMQArgs.SetURI(uri)
	}
	proxy.connectionArgs = rabbitMQArgs
}

func (proxy *RabbitMQConnectionProxy) GetConnection() (any, error) {
	args := proxy.connectionArgs
	conn, err := amqp.DialConfig(args.GetURI(), amqp.Config{
		Heartbeat: args.GetHeartbeat(),
		Dial:      amqp.DefaultDial(args.GetConnectionTimeout()),
	})
	return conn, err
}

func (proxy *RabbitMQConnectionProxy) Ping(connection any) error {
	if rabbitMQConn, ok := connection.(*amqp.Connection); ok {
		if rabbitMQConn.IsClosed() {
			return ErrConnectionIsClosed
		}
		return nil
	}
	return ErrConnectionType
}

func (proxy *RabbitMQConnectionProxy) IsNetworkErr(err error) bool {
	return connection_manager.IsNetworkErr(err) ||
			errors.As(err, amqp.ErrClosed) || 
			errors.As(err, amqp.ErrChannelMax)
}

func (proxy *RabbitMQConnectionProxy) CloseConnection(connection any) {
	if conn, ok := connection.(*amqp.Connection); ok {
		conn.Close()
	}
}
