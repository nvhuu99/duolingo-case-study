package connection_manager

import (
	"context"
	"sync"

	"github.com/google/uuid"
)

type ConnectionBuilder struct {
	proxy ConnectionProxy
	args  ConnectionArgs

	manager   *ConnectionManager
	managerMu sync.Mutex

	ctx context.Context
}

func NewConnectionBuilder(ctx context.Context) *ConnectionBuilder {
	return &ConnectionBuilder{
		ctx:  ctx,
		args: DefaultConnectionArgs(),
	}
}

func (builder *ConnectionBuilder) SetConnectionArgs(args ConnectionArgs) *ConnectionBuilder {
	builder.args = args
	if builder.proxy != nil {
		builder.proxy.SetArgsPanicIfInvalid(args)
	}
	return builder
}

func (builder *ConnectionBuilder) GetConnectionArgs() ConnectionArgs {
	return builder.args
}

func (builder *ConnectionBuilder) SetConnectionProxy(proxy ConnectionProxy) *ConnectionBuilder {
	proxy.SetArgsPanicIfInvalid(builder.GetConnectionArgs())
	builder.proxy = proxy
	return builder
}

func (builder *ConnectionBuilder) GetConnectionManager() *ConnectionManager {
	builder.managerMu.Lock()
	defer builder.managerMu.Unlock()
	if builder.manager == nil {
		builder.manager = &ConnectionManager{
			ctx:                 builder.ctx,
			connectionRetryWait: builder.args.GetRetryWait(),
			clients:             make(map[string]*Client),
			clientConnections:   make(map[string]any),
			connectionProxy:     builder.proxy,
		}
	}
	return builder.manager
}

func (builder *ConnectionBuilder) BuildClientAndRegisterToManager() *Client {
	manager := builder.GetConnectionManager()
	client := builder.CreateClient(uuid.NewString())
	manager.RegisterClient(client)
	return client
}

func (builder *ConnectionBuilder) CreateClient(id string) *Client {
	client := &Client{
		id:           id,
		ctx:          builder.ctx,
		readTimeout:  builder.args.GetReadTimeout(),
		writeTimeout: builder.args.GetWriteTimeout(),
		retryWait:    builder.args.GetRetryWait(),
	}
	return client
}

func (builder *ConnectionBuilder) Destroy() {
	builder.managerMu.Lock()
	defer builder.managerMu.Unlock()
	if builder.manager != nil {
		builder.manager.RemoveAllClients()
	}
	builder.manager = nil
}
