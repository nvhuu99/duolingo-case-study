package mongodb

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

type UserRepoBuilder struct {
	ctx                         context.Context
	hasUserRepoCreatedBefore    atomic.Bool
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

func NewUserRepoBuilder(ctx context.Context) *UserRepoBuilder {
	return &UserRepoBuilder{
		ctx:                ctx,
		connectionWait:     15 * time.Second,
		operationReadWait:  5 * time.Second,
		operationWriteWait: 10 * time.Second,
		operationRetryWait: 300 * time.Millisecond,
	}
}

func (builder *UserRepoBuilder) SetCredentials(user string, password string) *UserRepoBuilder {
	builder.user = url.QueryEscape(user)
	builder.password = url.QueryEscape(password)
	return builder
}

func (builder *UserRepoBuilder) SetHost(host string) *UserRepoBuilder {
	builder.host = host
	return builder
}

func (builder *UserRepoBuilder) SetPort(port string) *UserRepoBuilder {
	builder.port = port
	return builder
}

func (builder *UserRepoBuilder) SetOperationRetryWait(duration time.Duration) *UserRepoBuilder {
	connectionManager.connectionGraceWait = duration
	builder.operationRetryWait = duration
	return builder
}

func (builder *UserRepoBuilder) SetConnectionTimeOut(duration time.Duration) *UserRepoBuilder {
	builder.connectionWait = duration
	return builder
}

func (builder *UserRepoBuilder) SetOperationReadTimeOut(duration time.Duration) *UserRepoBuilder {
	builder.operationReadWait = duration
	return builder
}

func (builder *UserRepoBuilder) SetOperationWriteTimeOut(duration time.Duration) *UserRepoBuilder {
	builder.operationWriteWait = duration
	return builder
}

func (builder *UserRepoBuilder) SetDatabaseName(name string) *UserRepoBuilder {
	builder.databaseName = name
	return builder
}

func (builder *UserRepoBuilder) SetCollectionName(name string) *UserRepoBuilder {
	builder.collectionName = name
	return builder
}

func (builder *UserRepoBuilder) BuildConnectionManager() (*ConnectionManager, error) {
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

func (builder *UserRepoBuilder) BuildClientAndRegisterToManager() (*Client, error) {
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
	connectionManager.RegisterClient(client, true)
	return client, nil
}

func (builder *UserRepoBuilder) BuildRepo(client *Client) (*UserRepo, error) {
	if connectionManager == nil {
		return nil, ErrConnManagerHasNotCreated
	}
	defer builder.hasUserRepoCreatedBefore.Store(true)
	if err := builder.ensureUserRepoIsSingleton(); err != nil {
		return nil, err
	}
	return &UserRepo{client: client}, nil
}

func (builder *UserRepoBuilder) ensureUserRepoIsSingleton() error {
	if builder.hasUserRepoCreatedBefore.Load() {
		return ErrUserRepoSingletonViolation
	}
	return nil
}

func (builder *UserRepoBuilder) ensureConnManagerIsSingleton() error {
	if builder.hasConnManagerCreatedBefore.Load() {
		return ErrConnManagerSingletonViolation
	}
	return nil
}

func (builder *UserRepoBuilder) setDefaultArguments() {
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

func (builder *UserRepoBuilder) updateConnectionManagerUri() {
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
