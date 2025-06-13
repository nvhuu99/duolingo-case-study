package redis

import (
	"context"
	"errors"
	"fmt"
	"net"

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

func (proxy *RedisConnectionProxy) GetConnection() (any, error) {
	args := proxy.connectionArgs
	url := fmt.Sprintf("redis://%v:%v", args.GetHost(), args.GetPort())
	opt, err := redis_driver.ParseURL(url)
	if err != nil {
		return nil, err
	}
	return redis_driver.NewClient(opt), nil
}

func (proxy *RedisConnectionProxy) Ping(connection any) error {
	if redisConn, ok := connection.(*redis_driver.Client); ok {
		_, pingErr := redisConn.Ping(proxy.ctx).Result()
		return pingErr
	}
	return ErrConnectionType
}

func (proxy *RedisConnectionProxy) IsNetworkErr(err error) bool {
	var netErr net.Error
	return errors.As(err, &netErr)
}

func (proxy *RedisConnectionProxy) CloseConnection(connection any) {
	if redisConn, ok := connection.(*redis_driver.Client); ok {
		redisConn.Close()
	}
}
