package grpc_service

import (
	"context"
	"time"

	pb "duolingo/lib/log/grpc_service/proto/logger"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type LoggerGRPCClient struct {
	instance pb.LoggerClient
	ctx      context.Context
}

func NewLoggerGRPCClient(ctx context.Context, addr string) (*LoggerGRPCClient, error) {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	client := &LoggerGRPCClient{
		ctx:      ctx,
		instance: pb.NewLoggerClient(conn),
	}

	go func() {
		<-ctx.Done()
		conn.Close()
	}()

	return client, nil
}

func (client *LoggerGRPCClient) PushAll(lines [][]byte) error {
	ctx, cancel := context.WithTimeout(client.ctx, 10*time.Second)
	defer cancel()

	stream, err := client.instance.PushStream(ctx)
	if err != nil {
		return err
	}

	for _, line := range lines {
		if err := stream.Send(&pb.PushStreamRequest{Line: line}); err != nil {
			return err
		}
	}

	_, err = stream.CloseAndRecv()

	return err
}
