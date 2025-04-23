package grpc_service

import (
	lw "duolingo/lib/log/writer"
)

type LoggerService interface {
	Write(line *lw.Writable) error
}
