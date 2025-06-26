package buffer

import (
	"context"
	"sync"
	"time"
)

type BufferGroup[K comparable, T any] struct {
	limit       int
	interval    time.Duration
	consumeWait bool
	consumeFunc func(K, []T)

	groupMu sync.Mutex
	groups  map[K]*Buffer[T]
}

func NewBufferGroup[K comparable, T any]() *BufferGroup[K, T] {
	return &BufferGroup[K, T]{
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

func (gb *BufferGroup[K, T]) AddGroup(ctx context.Context, key K) *BufferGroup[K, T] {
	if gb.isAdded(key) {
		return gb
	}
	buf := NewBuffer[T]()
	buf.SetLimit(gb.limit).
		SetInterval(gb.interval).
		SetConsumeFunc(gb.consumeWait, func(t []T) {
			gb.consumeFunc(key, t)
		}).
		Start(ctx)

	gb.setGroup(key, buf)

	return gb
}

func (gb *BufferGroup[K, T]) RemoveGroup(key K) {
	if !gb.isAdded(key) {
		return
	}
	grp := gb.getGroup(key)

	gb.groupMu.Lock()
	delete(gb.groups, key)
	gb.groupMu.Unlock()

	grp.Stop()
}

func (gb *BufferGroup[K, T]) Stop() {
	gb.groupMu.Lock()
	defer gb.groupMu.Unlock()
	for key := range gb.groups {
		gb.groups[key].Stop()
	}
}

func (gb *BufferGroup[K, T]) Write(key K, items ...T) {
	if !gb.isAdded(key) {
		return
	}
	gb.getGroup(key).Write(items...)
}

func (gb *BufferGroup[K, T]) isAdded(key K) bool {
	gb.groupMu.Lock()
	defer gb.groupMu.Unlock()
	_, exist := gb.groups[key]
	return exist
}

func (gb *BufferGroup[K, T]) getGroup(key K) *Buffer[T] {
	gb.groupMu.Lock()
	defer gb.groupMu.Unlock()
	return gb.groups[key]
}

func (gb *BufferGroup[K, T]) setGroup(key K, buf *Buffer[T]) {
	gb.groupMu.Lock()
	defer gb.groupMu.Unlock()
	gb.groups[key] = buf
}
