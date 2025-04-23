package grpc_service

import (
	"context"
	"encoding/json"
	"io"
	"net"

	pb "duolingo/lib/log/grpc_service/proto/logger"
	lw "duolingo/lib/log/writer"

	"google.golang.org/grpc"
)

type LoggerGRPCServer struct {
	pb.UnimplementedLoggerServer

	addr   string
	ctx    context.Context
	logger LoggerService
}

func (srv *LoggerGRPCServer) UseLoggerService(logger LoggerService) {
	srv.logger = logger
}

func (srv *LoggerGRPCServer) Listen() error {
	server := grpc.NewServer()
	pb.RegisterLoggerServer(server, srv)

	go func() {
		<-srv.ctx.Done()
		server.GracefulStop()
	}()

	lis, err := net.Listen("tcp", srv.addr)
	if err != nil {
		return err
	}

	err = server.Serve(lis)

	return err
}

func (s *LoggerGRPCServer) PushStream(stream pb.Logger_PushStreamServer) error {
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			return stream.SendAndClose(&pb.PushStreamResponse{})
		}
		if err != nil {
			return err
		}
		line := new(lw.Writable)
		if err := json.Unmarshal(req.GetLine(), &line); err != nil {
			return err
		}
		if err := s.logger.Write(line); err != nil {
			return err
		}
	}
}
