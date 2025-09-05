package rabbitmq

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"duolingo/libraries/connection_manager"

	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitMQConnectionProxy struct {
	ctx            context.Context
	connectionArgs *RabbitMQConnectionArgs

	currentConnection *amqp.Connection
	connectionMu      sync.Mutex
}

func NewRabbitMQConnectionProxy(ctx context.Context) *RabbitMQConnectionProxy {
	proxy := &RabbitMQConnectionProxy{ctx: ctx}
	return proxy
}

/* Implement connection_manager.ConnectionProxy interface */

func (proxy *RabbitMQConnectionProxy) ConnectionName() string { return "RabbitMQ" }

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

func (proxy *RabbitMQConnectionProxy) MakeConnection() (any, error) {
	return proxy.getAMQPChannel()
}

func (proxy *RabbitMQConnectionProxy) Ping(connection any) error {
	if rabbitMQChan, ok := connection.(*amqp.Channel); ok {
		if rabbitMQChan.IsClosed() {
			return ErrConnectionIsClosed
		}
		return nil
	}
	return ErrConnectionType
}

func (proxy *RabbitMQConnectionProxy) IsNetworkErr(err error) bool {
	result := connection_manager.IsNetworkErr(err) ||
		errors.As(err, &amqp.ErrClosed) ||
		errors.As(err, &amqp.ErrChannelMax)

	return result
}

func (proxy *RabbitMQConnectionProxy) CloseConnection(connection any) {
	if ch, ok := connection.(*amqp.Channel); ok {
		ch.Close()
	}
}

func (proxy *RabbitMQConnectionProxy) getAMQPChannel() (*amqp.Channel, error) {
	conn, conErr := proxy.getAMQPConnection()
	if conErr != nil {
		return nil, conErr
	}
	args := proxy.connectionArgs
	// Declare channel
	ch, chErr := conn.Channel()
	if chErr != nil {
		return nil, chErr
	}
	// Declare quality of service
	qosErr := ch.Qos(
		int(args.GetPrefetchCount()),
		int(args.GetPrefetchLimit()),
		true, // Apply all channels
	)
	if qosErr != nil {
		return nil, qosErr
	}
	return ch, nil
}

func (proxy *RabbitMQConnectionProxy) getAMQPConnection() (
	*amqp.Connection,
	error,
) {
	proxy.connectionMu.Lock()
	defer proxy.connectionMu.Unlock()

	// Reuse the current connection
	if proxy.currentConnection != nil && !proxy.currentConnection.IsClosed() {
		return proxy.currentConnection, nil
	}

	// Recreate connection if closed
	args := proxy.connectionArgs
	newConn, connErr := amqp.DialConfig(args.GetURI(), amqp.Config{
		Heartbeat: args.GetHeartbeat(),
		Dial:      amqp.DefaultDial(args.GetConnectionTimeout()),
	})
	if connErr != nil {
		return nil, connErr
	}
	proxy.currentConnection = newConn

	return newConn, nil
}
