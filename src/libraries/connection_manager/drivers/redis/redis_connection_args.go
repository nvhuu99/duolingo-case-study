package redis

import (
	"duolingo/libraries/connection_manager"
)

type RedisConnectionArgs struct {
	connection_manager.ConnectionArgs

	Host     string
	Port     string
	User     string
	Password string
}
