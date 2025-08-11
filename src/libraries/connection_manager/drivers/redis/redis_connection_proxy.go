package redis

import (
	"context"
	"duolingo/libraries/connection_manager"
	"fmt"
	"strings"

	redis_driver "github.com/redis/go-redis/v9"
)

type RedisConnectionProxy struct {
	ctx            context.Context
	connectionArgs *RedisConnectionArgs
}

func NewRedisConnectionProxy(ctx context.Context) *RedisConnectionProxy {
	return &RedisConnectionProxy{ctx: ctx}
}

/* Implement connection_manager.ConnectionProxy interface */

func (proxy *RedisConnectionProxy) ConnectionName() string { return "Redis" }

func (proxy *RedisConnectionProxy) SetArgsPanicIfInvalid(args any) {
	redisArgs, ok := args.(*RedisConnectionArgs)
	if !ok {
		panic(ErrConnectionArgsType)
	}
	if redisArgs.GetHost() == "" || redisArgs.GetPort() == "" {
		panic(ErrInvalidConnectionArgs)
	}
	if redisArgs.GetURI() == "" {
		address := fmt.Sprintf("%v:%v", redisArgs.GetHost(), redisArgs.GetPort())
		credentials := fmt.Sprintf("%v:%v", redisArgs.GetUser(), redisArgs.GetPassword())
		uri := fmt.Sprintf("redis://%v/", address)
		if credentials != ":" {
			uri = fmt.Sprintf("redis://%v@%v/", credentials, address)
		}
		redisArgs.SetURI(uri)
	}
	proxy.connectionArgs = redisArgs
}

func (proxy *RedisConnectionProxy) MakeConnection() (any, error) {
	opt, err := redis_driver.ParseURL(proxy.connectionArgs.GetURI())
	if err != nil {
		return nil, err
	}

	redisClient := redis_driver.NewClient(opt)
	redisClient.AddHook(&EventEmitterHook{})

	return redisClient, nil
}

func (proxy *RedisConnectionProxy) Ping(connection any) error {
	if redisConn, ok := connection.(*redis_driver.Client); ok {
		_, pingErr := redisConn.Ping(proxy.ctx).Result()
		return pingErr
	}
	return ErrConnectionType
}

func (proxy *RedisConnectionProxy) IsNetworkErr(err error) bool {
	if err == nil {
		return false
	}
	mssg := err.Error()
	return connection_manager.IsNetworkErr(err) ||
		strings.Contains(mssg, "client is closed") ||
		strings.Contains(mssg, "connection pool exhausted") ||
		strings.Contains(mssg, "connection pool timeout")
}

func (proxy *RedisConnectionProxy) CloseConnection(connection any) {
	if redisConn, ok := connection.(*redis_driver.Client); ok {
		redisConn.Close()
	}
}
