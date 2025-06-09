package redis

import (
	"context"
	"errors"
	"fmt"
	"net"

	redis_driver "github.com/redis/go-redis/v9"
)

var (
	ErrConnectionType        = errors.New("provided with a connection not redis.Client")
	ErrConnectionArgsType    = errors.New("provided with an argument type that is not RedisConnectionArgs")
	ErrInvalidConnectionArgs = errors.New("provided with invalid connection arguments")
)

type RedisConnectionProxy struct {
	ctx            context.Context
	connectionArgs *RedisConnectionArgs
}

/* Implement connection_manager.ConnectionProxy interface */

func (proxy *RedisConnectionProxy) SetConnectionArgsWithPanicOnValidationErr(args any) {
	redisArgs, ok := args.(*RedisConnectionArgs)
	if !ok {
		panic(ErrConnectionArgsType)
	}
	if redisArgs.Host == "" || redisArgs.Port == "" {
		panic(ErrInvalidConnectionArgs)
	}
	if redisArgs.URI == "" {
		address := fmt.Sprintf("%v:%v", redisArgs.Host, redisArgs.Port)
		credentials := fmt.Sprintf("%v:%v", redisArgs.User, redisArgs.Password)
		uri := fmt.Sprintf("redis://%v/", address)
		if credentials != ":" {
			uri = fmt.Sprintf("redis://%v@%v/", credentials, address)
		}
		redisArgs.URI = uri
	}
	proxy.connectionArgs = redisArgs
}

func (proxy *RedisConnectionProxy) CreateConnection() (any, error) {
	args := proxy.connectionArgs
	opt, err := redis_driver.ParseURL(fmt.Sprintf("redis://%v:%v", args.Host, args.Port))
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

func (proxy *RedisConnectionProxy) IsNetworkError(err error) bool {
	var netErr net.Error
	return errors.As(err, &netErr)
}

func (proxy *RedisConnectionProxy) CloseConnection(connection any) {
	if redisConn, ok := connection.(*redis_driver.Client); ok {
		redisConn.Close()
	}
}
