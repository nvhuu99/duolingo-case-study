package buffer

import (
	"context"
	"time"
)

type BufferGroup[K comparable, T any] struct {
	groups      map[K]*Buffer[T]
	ctx         context.Context
	limit       int
	interval    time.Duration
	consumeWait bool
	consumeFunc func(K, []T)
}

func NewBufferGroup[K comparable, T any](ctx context.Context) *BufferGroup[K, T] {
	return &BufferGroup[K, T]{
		ctx:    ctx,
		groups: make(map[K]*Buffer[T]),
	}
}

func (gb *BufferGroup[K, T]) SetLimit(limit int) *BufferGroup[K, T] {
	gb.limit = limit
	return gb
}

func (gb *BufferGroup[K, T]) SetInterval(interval time.Duration) *BufferGroup[K, T] {
	gb.interval = interval
	return gb
}

func (gb *BufferGroup[K, T]) SetConsumeFunc(wait bool, consumeFunc func(K, []T)) *BufferGroup[K, T] {
	gb.consumeWait = wait
	gb.consumeFunc = consumeFunc
	return gb
}

func (gb *BufferGroup[K, T]) AddGroup(key K) *BufferGroup[K, T] {
	gb.groups[key] = NewBuffer[T](gb.ctx).
		SetLimit(gb.limit).
		SetInterval(gb.interval).
		SetConsumeFunc(gb.consumeWait, func(t []T) { gb.consumeFunc(key, t) })
	return gb
}

func (gb *BufferGroup[K, T]) RemoveGroup(key K) {
	if grp, exist := gb.groups[key]; exist {
		delete(gb.groups, key)
		grp.Stop()
	}
}

func (gb *BufferGroup[K, T]) Start() {
	for key := range gb.groups {
		gb.groups[key].Start()
	}
}

func (gb *BufferGroup[K, T]) Stop() {
	for key := range gb.groups {
		gb.groups[key].Stop()
	}
}

func (gb *BufferGroup[K, T]) Write(key K, items ...T) {
	if buffer, exist := gb.groups[key]; exist {
		buffer.Write(items...)
	}
}
