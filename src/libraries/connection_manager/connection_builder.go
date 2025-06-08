package connection_manager

import (
	"context"
	"errors"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
)

var (
	singletonManager            *ConnectionManager
	hasConnManagerCreatedBefore atomic.Bool

	ErrConnManagerSingletonViolation = errors.New("failed to build ConnectionManager due to singleton violation (build has already called)")
	ErrConnManagerHasNotCreated      = errors.New("ConnectionManager has been not created")
)

type ConnectionBuilder struct {
	ConnectionTimeout     time.Duration
	ConnectionRetryWait   time.Duration
	OperationReadTimeout  time.Duration
	OperationWriteTimeout time.Duration
	OperationRetryWait    time.Duration

	uri    string
	driver ConnectionProxy

	ctx context.Context
}

func NewConnectionBuilder(ctx context.Context) *ConnectionBuilder {
	return &ConnectionBuilder{
		ctx:                   ctx,
		ConnectionTimeout:     30 * time.Second,
		ConnectionRetryWait:   300 * time.Second,
		OperationReadTimeout:  10 * time.Second,
		OperationWriteTimeout: 10 * time.Second,
		OperationRetryWait:    300 * time.Millisecond,
	}
}

func (builder *ConnectionBuilder) SetConnectionDriver(driver ConnectionProxy) *ConnectionBuilder {
	builder.driver = driver
	return builder
}

func (builder *ConnectionBuilder) SetURI(uri string) *ConnectionBuilder {
	builder.uri = uri
	return builder
}

func (builder *ConnectionBuilder) SetConnectionTimeOut(duration time.Duration) *ConnectionBuilder {
	builder.ConnectionTimeout = duration
	return builder
}

func (builder *ConnectionBuilder) SetConnectionRetryWait(duration time.Duration) *ConnectionBuilder {
	builder.ConnectionRetryWait = duration
	return builder
}

func (builder *ConnectionBuilder) SetOperationReadTimeOut(duration time.Duration) *ConnectionBuilder {
	builder.OperationReadTimeout = duration
	return builder
}

func (builder *ConnectionBuilder) SetOperationWriteTimeOut(duration time.Duration) *ConnectionBuilder {
	builder.OperationWriteTimeout = duration
	return builder
}

func (builder *ConnectionBuilder) SetOperationRetryWait(duration time.Duration) *ConnectionBuilder {
	builder.OperationRetryWait = duration
	return builder
}

func (builder *ConnectionBuilder) BuildConnectionManager() (*ConnectionManager, error) {
	if singletonErr := builder.ensureConnManagerIsSingleton(); singletonErr != nil {
		return nil, singletonErr
	}
	defer hasConnManagerCreatedBefore.Store(true)

	singletonManager = &ConnectionManager{
		uri:                 builder.uri,
		ctx:                 builder.ctx,
		connectionTimeout:   builder.ConnectionTimeout,
		connectionRetryWait: builder.ConnectionRetryWait,
		clients:             make(map[string]*Client),
		clientConnections:   make(map[string]any),
		connectionProxy:     builder.driver,
	}

	return singletonManager, nil
}

func (builder *ConnectionBuilder) BuildClientAndRegisterToManager() (*Client, error) {
	if singletonManager == nil {
		return nil, ErrConnManagerHasNotCreated
	}
	client := builder.createClient(uuid.NewString())

	singletonManager.RegisterClient(client)

	return client, nil
}

func (builder *ConnectionBuilder) Destroy() {
	if singletonManager == nil {
		return
	}
	singletonManager.RemoveAllClients()
	singletonManager = nil
	hasConnManagerCreatedBefore.Store(false)
}

func (builder *ConnectionBuilder) ensureConnManagerIsSingleton() error {
	if hasConnManagerCreatedBefore.Load() {
		return ErrConnManagerSingletonViolation
	}
	return nil
}

func (builder *ConnectionBuilder) createClient(id string) *Client {
	client := &Client{
		id:                    id,
		ctx:                   builder.ctx,
		operationReadTimeout:  builder.OperationReadTimeout,
		operationWriteTimeout: builder.OperationWriteTimeout,
		operationRetryWait:    builder.OperationRetryWait,
	}
	return client
}
