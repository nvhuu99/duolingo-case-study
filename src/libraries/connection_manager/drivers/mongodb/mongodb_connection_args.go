package mongodb

import (
	"duolingo/libraries/connection_manager"
	"net/url"
)

type MongoConnectionArgs struct {
	*connection_manager.BaseConnectionArgs

	uri      string
	host     string
	port     string
	user     string
	password string
}

func DefaultMongoConnectionArgs() *MongoConnectionArgs {
	baseArgs := connection_manager.DefaultConnectionArgs()
	mongoArgs := &MongoConnectionArgs{
		BaseConnectionArgs: baseArgs,
		host:               "127.0.0.1",
		port:               "27017",
		user:               "",
		password:           "",
	}
	return mongoArgs
}

func (m *MongoConnectionArgs) GetURI() string {
	return m.uri
}

func (m *MongoConnectionArgs) SetURI(uri string) *MongoConnectionArgs {
	m.uri = uri
	return m
}

func (m *MongoConnectionArgs) GetHost() string {
	return m.host
}

func (m *MongoConnectionArgs) SetHost(host string) *MongoConnectionArgs {
	m.host = host
	return m
}

func (m *MongoConnectionArgs) GetPort() string {
	return m.port
}

func (m *MongoConnectionArgs) SetPort(port string) *MongoConnectionArgs {
	m.port = port
	return m
}

func (m *MongoConnectionArgs) GetUser() string {
	return m.user
}

func (m *MongoConnectionArgs) SetUser(user string) *MongoConnectionArgs {
	m.user = url.QueryEscape(user)
	return m
}

func (m *MongoConnectionArgs) GetPassword() string {
	return m.password
}

func (m *MongoConnectionArgs) SetPassword(password string) *MongoConnectionArgs {
	m.password = url.QueryEscape(password)
	return m
}

func (m *MongoConnectionArgs) SetCredentials(user string, password string) *MongoConnectionArgs {
	m.user = url.QueryEscape(user)
	m.password = url.QueryEscape(password)
	return m
}
