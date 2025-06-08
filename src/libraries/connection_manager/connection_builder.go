package connection_manager

import (
	"context"
	"errors"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
)

var (
	connectionManager *ConnectionManager

	ErrConnManagerSingletonViolation = errors.New("failed to build ConnectionManager due to singleton violation (build has already called)")
	ErrConnManagerHasNotCreated      = errors.New("ConnectionManager has been not created")
)

type ConnectionBuilder struct {
	ctx                         context.Context
	hasConnManagerCreatedBefore atomic.Bool

	uri string
	connectionWait     time.Duration
	operationReadWait  time.Duration
	operationWriteWait time.Duration
	operationRetryWait time.Duration
	driver ConnectionProxy
}

func NewConnectionBuilder(ctx context.Context) *ConnectionBuilder {
	return &ConnectionBuilder{
		ctx:                ctx,
		connectionWait:     15 * time.Second,
		operationReadWait:  5 * time.Second,
		operationWriteWait: 10 * time.Second,
		operationRetryWait: 300 * time.Millisecond,
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

func (builder *ConnectionBuilder) SetOperationRetryWait(duration time.Duration) *ConnectionBuilder {
	builder.operationRetryWait = duration
	return builder
}

func (builder *ConnectionBuilder) SetConnectionTimeOut(duration time.Duration) *ConnectionBuilder {
	builder.connectionWait = duration
	return builder
}

func (builder *ConnectionBuilder) SetOperationReadTimeOut(duration time.Duration) *ConnectionBuilder {
	builder.operationReadWait = duration
	return builder
}

func (builder *ConnectionBuilder) SetOperationWriteTimeOut(duration time.Duration) *ConnectionBuilder {
	builder.operationWriteWait = duration
	return builder
}

func (builder *ConnectionBuilder) BuildConnectionManager() (*ConnectionManager, error) {
	if singletonErr := builder.ensureConnManagerIsSingleton(); singletonErr != nil {
		return nil, singletonErr
	}
	defer builder.hasConnManagerCreatedBefore.Store(true)
	connectionManager = &ConnectionManager{
		uri: builder.uri,
		ctx:                 builder.ctx,
		connectionGraceWait: builder.connectionWait,
		clients:             make(map[string]*Client),
		clientConnections:   make(map[string]any),
		connectionProxy: builder.driver,
	}
	builder.registerManagerSpecialClient()
	return connectionManager, nil
}

func (builder *ConnectionBuilder) BuildClientAndRegisterToManager() (*Client, error) {
	if connectionManager == nil {
		return nil, ErrConnManagerHasNotCreated
	}
	client := builder.createClient(uuid.NewString())
	connectionManager.RegisterClient(client)
	return client, nil
}

func (builder *ConnectionBuilder) Destroy() {
	if connectionManager == nil {
		return
	}
	connectionManager.RemoveAllClients()
	connectionManager = nil
	builder.hasConnManagerCreatedBefore.Store(false)
}


func (builder *ConnectionBuilder) ensureConnManagerIsSingleton() error {
	if builder.hasConnManagerCreatedBefore.Load() {
		return ErrConnManagerSingletonViolation
	}
	return nil
}

func (builder *ConnectionBuilder) registerManagerSpecialClient() {
	if connectionManager == nil {
		return
	}
	client := builder.createClient(managerSpecialClientId)
	connectionManager.RegisterClient(client)
}

func (builder *ConnectionBuilder) createClient(id string) *Client {
	client := &Client{
		id:                 id,
		ctx:                builder.ctx,
		operationReadWait:  builder.operationReadWait,
		operationWriteWait: builder.operationWriteWait,
		operationRetryWait: builder.operationRetryWait,
	}
	return client
}