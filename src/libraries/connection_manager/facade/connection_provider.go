package facade

import (
	"context"
	"sync"

	"duolingo/libraries/connection_manager/drivers/mongodb"
	"duolingo/libraries/connection_manager/drivers/redis"
)

var (
	provider           *ConnectionProvider
	providerCreateOnce sync.Once
)

type ConnectionProvider struct {
	redisBuilder    *redis.RedisConnectionBuilder
	redisCreateOnce sync.Once

	mongoBuilder    *mongodb.MongoConnectionBuilder
	mongoCreateOnce sync.Once

	ctx context.Context
}

func Provider(ctx context.Context) *ConnectionProvider {
	providerCreateOnce.Do(func() {
		if ctx == nil {
			panic("fail to create ConnectionProvier using nil context")
		}
		provider = &ConnectionProvider{ctx: ctx}
	})
	return provider
}

/* Redis provider methods */

func (p *ConnectionProvider) InitRedisWithBasicArgs(
	host string,
	port string,
) *ConnectionProvider {
	args := redis.DefaultRedisConnectionArgs().SetHost(host).SetPort(port)
	p.InitRedis(args)
	return p
}

func (p *ConnectionProvider) InitRedis(
	args *redis.RedisConnectionArgs,
) *ConnectionProvider {
	p.redisCreateOnce.Do(func() {
		p.redisBuilder = redis.NewRedisConnectionBuilder(p.ctx, args)
	})
	return p
}

func (p *ConnectionProvider) GetRedisClient() *redis.RedisClient {
	return p.redisBuilder.BuildClientAndRegisterToManager()
}

/* Mongo provider methods */

func (p *ConnectionProvider) InitMongoWithBasicArgs(
	host string,
	usr string,
	pwd string,
) *ConnectionProvider {
	args := mongodb.DefaultMongoConnectionArgs()
	args.SetHost(host)
	args.SetPort("27017")
	args.SetCredentials(usr, pwd)
	p.InitMongo(args)
	return p
}

func (p *ConnectionProvider) InitMongo(
	args *mongodb.MongoConnectionArgs,
) *ConnectionProvider {
	p.mongoCreateOnce.Do(func() {
		p.mongoBuilder = mongodb.NewMongoConnectionBuilder(p.ctx, args)
	})
	return p
}

func (p *ConnectionProvider) GetMongoClient() *mongodb.MongoClient {
	return p.mongoBuilder.BuildClientAndRegisterToManager()
}
