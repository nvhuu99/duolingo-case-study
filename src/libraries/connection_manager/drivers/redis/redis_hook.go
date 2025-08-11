package redis

import (
	"context"
	"net"

	events "duolingo/libraries/events/facade"

	"github.com/redis/go-redis/v9"
)

type EventEmitterHook struct{}

func (EventEmitterHook) DialHook(next redis.DialHook) redis.DialHook {
	return func(ctx context.Context, network, addr string) (net.Conn, error) {
		return next(ctx, network, addr)
	}
}

func (EventEmitterHook) ProcessHook(next redis.ProcessHook) redis.ProcessHook {
	return func(ctx context.Context, cmd redis.Cmder) error {
		var err error

		evt := events.Start(ctx, "redis.execute_command", map[string]any{
			"redis.command.name": cmd.Name(),
		})
		defer events.End(evt, true, err, nil)

		err = next(ctx, cmd)

		return err
	}
}

func (EventEmitterHook) ProcessPipelineHook(next redis.ProcessPipelineHook) redis.ProcessPipelineHook {
	return func(ctx context.Context, cmds []redis.Cmder) error {
		var err error

		evt := events.Start(ctx, "redis.execute_pipeline", map[string]any{
			"redis.command.name": "pipeline",
		})
		defer events.End(evt, true, err, nil)

		err = next(ctx, cmds)

		return err
	}
}
