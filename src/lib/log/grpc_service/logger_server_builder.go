package grpc_service

import (
	"context"
	mongo "duolingo/lib/log/driver/grpc_service/mongo"
)

type LoggerGRPCServerBuilder struct {
	srv *LoggerGRPCServer
	err error
}

func NewLoggerGRPCServerBuilder(ctx context.Context, addr string) *LoggerGRPCServerBuilder {
	return &LoggerGRPCServerBuilder{
		srv: &LoggerGRPCServer{
			ctx:  ctx,
			addr: addr,
		},
	}
}

func (builder *LoggerGRPCServerBuilder) UseMongoLoggerService(uri string, db string, coll string) *LoggerGRPCServerBuilder {
	mongoService, err := mongo.NewMongoLogService(builder.srv.ctx, uri, db, coll)
	if err != nil {
		builder.err = err
	} else {
		builder.srv.UseLoggerService(mongoService)
	}
	return builder
}

func (builder *LoggerGRPCServerBuilder) GetServer() (*LoggerGRPCServer, error) {
	if builder.err != nil {
		return nil, builder.err
	}
	return builder.srv, nil
}
