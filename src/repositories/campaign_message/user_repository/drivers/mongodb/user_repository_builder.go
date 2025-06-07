package mongodb

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"sync/atomic"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
)

var (
	connectionManager     *ConnectionManager
	errSingletonViolation = errors.New("failed to build UserRepo due to singleton violation (build has already called)")
)

const (
	DEFAUTL_COLLECTION_NAME = "users"
	DEFAULT_DATABASE_NAME   = "duolingo"
	DEFAULT_HOST            = "localhost"
	DEFAULT_PORT            = "27017" // standard default mongodb port
)

type UserRepoBuilder struct {
	ctx                      context.Context
	hasUserRepoCreatedBefore atomic.Bool

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
	if connectionManager == nil {
		connectionManager = &ConnectionManager{
			ctx:                 ctx,
			connectionGraceWait: 300 * time.Millisecond,
			clients:             make(map[string]*Client),
			clientConnections:   make(map[string]*mongo.Client),
		}
	}
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

func (builder *UserRepoBuilder) Build() (*UserRepo, error) {
	if singletonErr := builder.ensureUserRepoIsSingleton(); singletonErr != nil {
		return nil, singletonErr
	}
	builder.setDefaultArguments()
	builder.updateConnectionManagerUri()
	repo := builder.createUserRepoAndRegisterToManager()
	return repo, nil
}

func (builder *UserRepoBuilder) ensureUserRepoIsSingleton() error {
	if builder.hasUserRepoCreatedBefore.Load() {
		builder.hasUserRepoCreatedBefore.Store(true)
		return errSingletonViolation
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

func (builder *UserRepoBuilder) createUserRepoAndRegisterToManager() *UserRepo {
	client := &Client{
		ctx:                builder.ctx,
		databaseName:       builder.databaseName,
		collectionName:     builder.collectionName,
		connectionWait:     builder.connectionWait,
		operationReadWait:  builder.operationReadWait,
		operationWriteWait: builder.operationWriteWait,
		operationRetryWait: builder.operationRetryWait,
	}
	connectionManager.RegisterClient(client, true)
	return &UserRepo{
		client: client,
	}
}
