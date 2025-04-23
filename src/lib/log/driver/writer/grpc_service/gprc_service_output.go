package grpc_service

import (
	"context"
	grpc "duolingo/lib/log/grpc_service"
	wt "duolingo/lib/log/writer"
	"encoding/json"
)

type GRPCServiceOutput struct {
	client *grpc.LoggerGRPCClient
}

func NewGRPCServiceOutput(ctx context.Context, addr string) (*GRPCServiceOutput, error) {
	output := new(GRPCServiceOutput)
	client, err := grpc.NewLoggerGRPCClient(ctx, addr)
	if err != nil {
		return nil, err
	}
	output.client = client
	return output, nil
}

func (output *GRPCServiceOutput) Flush(items []*wt.Writable) error {
	lines := make([][]byte, len(items))
	for i, item := range items {
		line, err := json.Marshal(item)
		if err != nil {
			return err
		}
		lines[i] = line
	}

	err := output.client.PushAll(lines)

	return err
}
