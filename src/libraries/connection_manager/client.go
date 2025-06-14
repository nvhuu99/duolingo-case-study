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

	id           string
	readTimeout  time.Duration
	writeTimeout time.Duration
	retryWait    time.Duration

	ctx context.Context
}

func (client *Client) GetClientId() string {
	return client.id
}

func (client *Client) GetReadTimeout() time.Duration {
	return client.readTimeout
}

func (client *Client) GetWriteTimeout() time.Duration {
	return client.writeTimeout
}

func (client *Client) GetDefaultTimeOut() time.Duration {
	return max(client.writeTimeout, client.readTimeout)
}

func (client *Client) GetRetryWait() time.Duration {
	return client.retryWait
}

func (client *Client) GetConnection() any {
	return client.connectionManager.GetClientConnection(client)
}

func (client *Client) IsNetworkErr(err error) bool {
	return client.connectionManager.IsNetworkErr(err)
}

func (client *Client) ExecuteClosure(
	timeout time.Duration,
	closure func(ctx context.Context, connection any) error,
) error {
	timeoutCtx, timeoutCancel := context.WithTimeout(client.ctx, timeout)
	defer timeoutCancel()

	done := make(chan bool, 1)
	errChan := make(chan error, 1)
	go func() {
		defer func() {
			close(done)
			close(errChan)
		}()
		client.executeClosureWithRetryOnNetworkErr(timeoutCtx, closure, errChan)
		done <- true
	}()

	for {
		select {
		case <-client.ctx.Done():
			return ErrClientContextCanceled
		case <-timeoutCtx.Done():
			return ErrClientOperationTimeout
		case <-done:
			if len(errChan) > 0 {
				return <-errChan
			}
			return nil
		}
	}
}

func (client *Client) executeClosureWithRetryOnNetworkErr(
	timeoutCtx context.Context,
	closure func(ctx context.Context, connection any) error,
	errChan chan error,
) {
	for {
		select {
		case <-timeoutCtx.Done():
			return
		default:
			// retry getting connection
			conn := client.connectionManager.GetClientConnection(client)
			if conn == nil {
				time.Sleep(client.retryWait)
				continue
			}
			// retry on network err
			err := closure(timeoutCtx, conn)
			if client.IsNetworkErr(err) {
				time.Sleep(client.retryWait)
				continue
			}
			// exit normally
			if err != nil {
				errChan <- err
			}
			return
		}
	}
}
