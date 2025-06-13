package redis

import "errors"

var (
	ErrConnectionType        = errors.New("provided with a connection not redis.Client")
	ErrConnectionArgsType    = errors.New("provided with an argument type that is not RedisConnectionArgs")
	ErrInvalidConnectionArgs = errors.New("provided with invalid connection arguments")
)
