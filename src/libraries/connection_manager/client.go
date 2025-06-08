package connection_manager

import (
	"context"
	"errors"
	"time"
)

var (
	ErrClientOperationTimeout = errors.New("client operation failed due to timeout exceeded")
	ErrClientContextCanceled  = errors.New("client operation canceled by context")
	ErrClientPingFailure      = errors.New("client ping failed due to network failure")
)

type Client struct {
	connectionManager *ConnectionManager

	id                    string
	operationReadTimeout  time.Duration
	operationWriteTimeout time.Duration
	operationRetryWait    time.Duration

	ctx context.Context
}

func (client *Client) GetClientId() string {
	return client.id
}

func (client *Client) GetReadTimeout() time.Duration {
	return client.operationReadTimeout
}

func (client *Client) GetWriteTimeout() time.Duration {
	return client.operationWriteTimeout
}

func (client *Client) GetDefaultTimeOut() time.Duration {
	return max(client.operationWriteTimeout, client.operationReadTimeout)
}

func (client *Client) GetRetryWait() time.Duration {
	return client.operationRetryWait
}

func (client *Client) GetConnection() any {
	return client.connectionManager.GetClientConnection(client)
}

func (client *Client) ExecuteClosure(
	wait time.Duration,
	closure func(ctx context.Context, connection any) error,
) error {
	timeoutCtx, timeoutCancel := context.WithTimeout(client.ctx, wait)
	defer timeoutCancel()

	done := make(chan bool, 1)
	go client.executeClosureWithRetryOnNetworkErr(timeoutCtx, done, closure)

	for {
		select {
		case <-client.ctx.Done():
			return ErrClientContextCanceled
		case <-timeoutCtx.Done():
			return ErrClientOperationTimeout
		case <-done:
			return nil
		}
	}
}

func (client *Client) executeClosureWithRetryOnNetworkErr(
	timeoutCtx context.Context,
	done chan bool,
	closure func(ctx context.Context, connection any) error,
) {
	defer func() {
		done <- true
	}()
	for {
		select {
		case <-timeoutCtx.Done():
			return
		default:
			if conn := client.connectionManager.GetClientConnection(client); conn != nil {
				err := closure(timeoutCtx, conn)
				if err == nil || !client.connectionManager.IsNetworkError(err) {
					return // exit normally
				}
			}
			time.Sleep(client.operationRetryWait)
			continue
		}
	}
}
