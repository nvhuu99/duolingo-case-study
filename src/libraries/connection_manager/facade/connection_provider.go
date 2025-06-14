package facade

import (
	"context"
	"sync"

	"duolingo/libraries/connection_manager/drivers/mongodb"
	"duolingo/libraries/connection_manager/drivers/redis"
	"duolingo/libraries/connection_manager/drivers/rabbitmq"
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

	rabbitMQBuilder    *rabbitmq.RabbitMQConnectionBuilder
	rabbitMQCreateOnce sync.Once

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

func (p *ConnectionProvider) InitRedisWithBasicArgs(host string) *ConnectionProvider {
	args := redis.DefaultRedisConnectionArgs().SetHost(host)
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

/* RabbitMQ provider methods */

func (p *ConnectionProvider) InitRabbitMQWithBasicArgs(
	host string,
	usr string,
	pwd string,
) *ConnectionProvider {
	args := rabbitmq.DefaultRabbitMQConnectionArgs()
	args.SetHost(host)
	args.SetCredentials(usr, pwd)
	p.InitRabbitMQ(args)
	return p
}

func (p *ConnectionProvider) InitRabbitMQ(
	args *rabbitmq.RabbitMQConnectionArgs,
) *ConnectionProvider {
	p.rabbitMQCreateOnce.Do(func() {
		p.rabbitMQBuilder = rabbitmq.NewRabbitMQConnectionBuilder(p.ctx, args)
	})
	return p
}

func (p *ConnectionProvider) GetRabbitMQClient() *rabbitmq.RabbitMQClient {
	return p.rabbitMQBuilder.BuildClientAndRegisterToManager()
}
