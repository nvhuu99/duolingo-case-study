package mongodb

import (
	"context"
	"errors"
	"sync/atomic"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
)

var (
	ErrOperationTimeout = errors.New("client operation failed due to timeout exceeded")
	ErrContextCanceled  = errors.New("client operation canceled by context")
)

type Client struct {
	ctx                context.Context
	reconnectTriggered atomic.Bool

	databaseName      string
	collectionName    string
	connectionManager *ClientConnectionManager

	connectionWait     time.Duration
	operationReadWait  time.Duration
	operationWriteWait time.Duration
	operationRetryWait time.Duration
}

func (client *Client) GetClientId() string {
	return ""
}

func (client *Client) GetReadTimeout() time.Duration {
	return client.operationReadWait
}

func (client *Client) GetWriteTimeout() time.Duration {
	return client.operationWriteWait
}

func (client *Client) GetDefaultTimeOut() time.Duration {
	return max(client.operationWriteWait, client.operationReadWait)
}

func (client *Client) GetConnectionTimeOut() time.Duration {
	return client.connectionWait
}

func (client *Client) ExecuteClosure(
	wait time.Duration,
	closure func(ctx context.Context, collection *mongo.Collection) error,
) (timeoutErr error) {
	done := make(chan bool, 1)
	timeoutCtx, timeoutCancel := context.WithTimeout(client.ctx, wait)
	defer timeoutCancel()

	go client.executeClosureWithRetryOnNetworkErr(timeoutCtx, done, closure)

	for {
		select {
		case <-client.ctx.Done():
			return ErrContextCanceled
		case <-timeoutCtx.Done():
			return ErrOperationTimeout
		case <-done:
			return nil
		}
	}
}

func (client *Client) executeClosureWithRetryOnNetworkErr(
	ctx context.Context,
	done chan bool,
	closure func(ctx context.Context, collection *mongo.Collection) error,
) {
	defer func() { done <- true }()
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}
		collection, collectionErr := client.getCollection()
		if collectionErr != nil {
			time.Sleep(client.operationRetryWait)
			continue
		}
		if err := closure(ctx, collection); mongo.IsNetworkError(err) {
			client.ensureSingleReconnectAttempt() // thread-safe reconnect attempt
			time.Sleep(client.operationRetryWait)
			continue
		}
		// not return the closure() error,
		// since executeClosureWithRetryOnNetworkErr() only cares for network errors
		// other error is the external user's responsibility
		return
	}
}

func (client *Client) ensureSingleReconnectAttempt() {
	if client.reconnectTriggered.Load() {
		return
	}
	client.reconnectTriggered.Store(true)
	client.getConnection(true) // trigger reconnection
	client.reconnectTriggered.Store(false)
}

func (client *Client) getConnection(forceReconnect bool) (*mongo.Client, error) {
	conn, err := client.connectionManager.GetClientConnection(client, forceReconnect)
	return conn, err
}

func (client *Client) getCollection() (*mongo.Collection, error) {
	conn, err := client.getConnection(false)
	if err != nil {
		return nil, err
	}
	return conn.Database(client.databaseName).Collection(client.collectionName), nil
}
