package mongo_connect

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/mongo"
)

var (
	connectionManager *ConnectionManager

	ErrUserRepoSingletonViolation    = errors.New("failed to build UserRepo due to singleton violation (build has already called)")
	ErrConnManagerSingletonViolation = errors.New("failed to build ConnectionManager due to singleton violation (build has already called)")
	ErrConnManagerHasNotCreated      = errors.New("ConnectionManager has been not created")
)

const (
	DEFAUTL_COLLECTION_NAME = "users"
	DEFAULT_DATABASE_NAME   = "duolingo"
	DEFAULT_HOST            = "localhost"
	DEFAULT_PORT            = "27017" // standard default mongodb port
)

type ConnectionBuilder struct {
	ctx                         context.Context
	hasConnManagerCreatedBefore atomic.Bool

	host               string
	port               string
	user               string
	password           string
	databaseName       string
	collectionName     string
	connectionWait     time.Duration
	operationReadWait  time.Duration
	operationWriteWait time.Duration
	operationRetryWait time.Duration
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

func (builder *ConnectionBuilder) SetCredentials(user string, password string) *ConnectionBuilder {
	builder.user = url.QueryEscape(user)
	builder.password = url.QueryEscape(password)
	return builder
}

func (builder *ConnectionBuilder) SetHost(host string) *ConnectionBuilder {
	builder.host = host
	return builder
}

func (builder *ConnectionBuilder) SetPort(port string) *ConnectionBuilder {
	builder.port = port
	return builder
}

func (builder *ConnectionBuilder) SetOperationRetryWait(duration time.Duration) *ConnectionBuilder {
	connectionManager.connectionGraceWait = duration
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

func (builder *ConnectionBuilder) SetDatabaseName(name string) *ConnectionBuilder {
	builder.databaseName = name
	return builder
}

func (builder *ConnectionBuilder) SetCollectionName(name string) *ConnectionBuilder {
	builder.collectionName = name
	return builder
}

func (builder *ConnectionBuilder) BuildConnectionManager() (*ConnectionManager, error) {
	if singletonErr := builder.ensureConnManagerIsSingleton(); singletonErr != nil {
		return nil, singletonErr
	}
	defer builder.hasConnManagerCreatedBefore.Store(true)
	connectionManager = &ConnectionManager{
		ctx:                 builder.ctx,
		connectionGraceWait: 300 * time.Millisecond,
		clients:             make(map[string]*Client),
		clientConnections:   make(map[string]*mongo.Client),
	}
	builder.setDefaultArguments()
	builder.updateConnectionManagerUri()
	return connectionManager, nil
}

func (builder *ConnectionBuilder) BuildClientAndRegisterToManager() (*Client, error) {
	if connectionManager == nil {
		return nil, ErrConnManagerHasNotCreated
	}
	client := &Client{
		id:                 uuid.NewString(),
		ctx:                builder.ctx,
		databaseName:       builder.databaseName,
		collectionName:     builder.collectionName,
		connectionWait:     builder.connectionWait,
		operationReadWait:  builder.operationReadWait,
		operationWriteWait: builder.operationWriteWait,
		operationRetryWait: builder.operationRetryWait,
	}
	connectionManager.RegisterClient(client, false)
	return client, nil
}

func (builder *ConnectionBuilder) ensureConnManagerIsSingleton() error {
	if builder.hasConnManagerCreatedBefore.Load() {
		return ErrConnManagerSingletonViolation
	}
	return nil
}

func (builder *ConnectionBuilder) setDefaultArguments() {
	if builder.databaseName == "" {
		builder.collectionName = DEFAULT_DATABASE_NAME
	}
	if builder.collectionName == "" {
		builder.collectionName = DEFAUTL_COLLECTION_NAME
	}
	if builder.port == "" {
		builder.port = DEFAULT_PORT
	}
	if builder.host == "" {
		builder.port = DEFAULT_HOST
	}
}

func (builder *ConnectionBuilder) updateConnectionManagerUri() {
	address := fmt.Sprintf("%v:%v", builder.host, builder.port)
	credentials := fmt.Sprintf("%v:%v", builder.user, builder.password)
	uri := fmt.Sprintf("mongodb://%v/", address)
	if credentials != ":" {
		uri = fmt.Sprintf("mongodb://%v@%v/", address, credentials)
	}
	if connectionManager.uri != uri {
		connectionManager.uri = uri
	}
}

func (builder *ConnectionBuilder) Destroy() {
	if connectionManager != nil {
		connectionManager.DiscardClients()
		connectionManager = nil
	}
}
