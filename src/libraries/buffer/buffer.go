package buffer

import (
	"context"
	"sync/atomic"
	"time"
)

type Buffer[T any] struct {
	interval    time.Duration
	limit       int
	consumeFunc func(context.Context, []T)
	consumeWait bool

	bufferCh chan T
	writeCh  chan T
	buffered atomic.Int32
	started  atomic.Bool
	flushing atomic.Bool

	bufferCtx    context.Context
	bufferCancel context.CancelFunc
}

func NewBuffer[T any]() *Buffer[T] {
	return &Buffer[T]{
		limit:    1000,
		interval: 2 * time.Second,
	}
}

func (b *Buffer[T]) SetLimit(limit int) *Buffer[T] {
	b.limit = limit
	return b
}

func (b *Buffer[T]) SetInterval(interval time.Duration) *Buffer[T] {
	b.interval = interval
	return b
}

func (b *Buffer[T]) SetConsumeFunc(
	wait bool,
	consumeFunc func(context.Context, []T),
) *Buffer[T] {
	b.consumeFunc = consumeFunc
	b.consumeWait = wait
	return b
}

func (b *Buffer[T]) Start(ctx context.Context) {
	if b.started.Load() {
		return
	}
	defer b.started.Store(true)

	b.bufferCh = make(chan T, b.limit)
	b.writeCh = make(chan T, b.limit)
	b.bufferCtx, b.bufferCancel = context.WithCancel(ctx)

	go b.run()
}

func (b *Buffer[T]) Stop() {
	if b.started.Load() {
		b.started.Store(false)
		b.bufferCancel()
	}
}

func (b *Buffer[T]) Write(items ...T) {
	if b.started.Load() {
		for i := range items {
			b.writeCh <- items[i]
		}
	}
}

func (b *Buffer[T]) Size() int {
	return int(b.buffered.Load())
}

func (b *Buffer[T]) Flush() {
	if !b.started.Load() || b.flushing.Load() {
		return
	}
	b.flushing.Store(true)
	defer b.flushing.Store(false)

	count := b.buffered.Load()
	items := make([]T, count)
	for i := range count {
		items[i] = <-b.bufferCh
		b.buffered.Add(-1)
	}

	b.callConsumeFunc(b.bufferCtx, items)
}

func (b *Buffer[T]) run() {
	flushInterval := time.NewTicker(b.interval)
	for {
		select {
		case <-b.bufferCtx.Done():
			go b.Flush()
			return
		case <-flushInterval.C:
			go b.Flush()
		case item := <-b.writeCh:
			b.buffer(item)
			if b.checkLimit() {
				go b.Flush()
			}
		}
	}
}

func (b *Buffer[T]) buffer(item T) {
	b.bufferCh <- item
	b.buffered.Add(1)
}

func (b *Buffer[T]) checkLimit() bool {
	return b.buffered.Load() >= int32(b.limit)
}

func (b *Buffer[T]) callConsumeFunc(ctx context.Context, items []T) {
	if len(items) == 0 {
		return
	}
	if b.consumeWait {
		b.consumeFunc(ctx, items)
	} else {
		go b.consumeFunc(ctx, items)
	}
}
