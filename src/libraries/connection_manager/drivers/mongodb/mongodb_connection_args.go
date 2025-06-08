package mongodb

import (
	"duolingo/libraries/connection_manager"
)

type MongoConnectionArgs struct {
	connection_manager.ConnectionArgs

	Host     string
	Port     string
	User     string
	Password string
}
