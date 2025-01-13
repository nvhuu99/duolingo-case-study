package redismanager

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strconv"
	"time"

	bm "duolingo/lib/batch-manager"
	redis "github.com/redis/go-redis/v9"
)

type RedisBatchManager struct {
	rdb *redis.Client

	name	   string
	start      int
	end        int
	size       int
	len		   int

	ctx		   context.Context
	timeout    time.Duration
}

func GetBatchManager(ctx context.Context, name string, start int, end int, size int) *RedisBatchManager {
	if start > end || size == 0 || size > (end - start + 1) {
		panic("batch manager: batch manager arguments invalid")
	}
	
	m := RedisBatchManager{}
	m.ctx = ctx
	m.start = start
	m.end = end
	m.size = size
	m.name = name
	m.len = int(math.Ceil(float64(end - start + 1) / float64(size))) 
	m.timeout = 5 * time.Second
	
	return &m
}

func (m *RedisBatchManager) SetConnection(host string, port string) error {
	opt, err := redis.ParseURL(fmt.Sprintf("redis://%v:%v", host, port))
	if err != nil {
		return err
	}
	m.rdb = redis.NewClient(opt)
	
	return nil
}

func (m *RedisBatchManager) Reset() error {
	key := fmt.Sprintf("batch_manager:%v:batch_ids", m.name)
	batchIds, err := m.rdb.SMembers(m.ctx, key).Result()
	if err != nil && err != redis.Nil {
		return err
	}

	for _, id := range batchIds {
		// clear all locks
		for _, l := range m.lock(id, "idx", "available", "batch") {
			err := m.rdb.Del(m.ctx, l).Err()
			if err != nil && err != redis.Nil {
				return err
			}
		}
		// clear all values
		err := m.rdb.Del(m.ctx, m.key(id, "idx"), m.key(id, "available"), m.key(id, "assigned")).Err()
		if err != nil && err != redis.Nil {
			return err
		}
	}

	return nil
}

func (m *RedisBatchManager) NewBatch(id string) error {
	key := fmt.Sprintf("batch_manager:%v:batch_ids", m.name)
	lock := strconv.Itoa(int(time.Now().UnixMilli()))
	acquireLock(m.ctx, m.rdb, m.timeout, lock, key)
	defer releaseLock(m.ctx, m.rdb, lock, key)

	err := m.rdb.SAdd(m.ctx, key, id).Err()
	if err != nil && err != redis.Nil {
		return err
	}
	
	return nil
}

func (m *RedisBatchManager) Progress(id string, itemId string, val int) error {
	// Check if batch id exists
	exists, err := m.rdb.HExists(m.ctx, m.key(id, "assigned"), itemId).Result()
	if err != nil {
		return err
	}
	if !exists {
		return errors.New("batch manager: can not update progress for an unassigned batch")
	}
	// Get batch
	var batch bm.BatchItem
	result, err := m.rdb.HGet(m.ctx, m.key(id, "assigned"), itemId).Result()
	if err != nil && err != redis.Nil {
		return err
	}
	json.Unmarshal([]byte(result), &batch)
	// Update batch progress
	batch.Progress = val
	bJson, _ := json.Marshal(batch)
	err = m.rdb.HSet(m.ctx, m.key(id, "assigned"), itemId, string(bJson)).Err()

	return err
}

func (m *RedisBatchManager) Next(id string) (*bm.BatchItem, error) {
	// Get next batch start & end 
	lock := strconv.Itoa(int(time.Now().UnixMilli()))
	acquireLock(m.ctx, m.rdb, m.timeout, lock, m.lock(id, "idx", "available")...)
	next, err := m.available(id)
	releaseLock(m.ctx, m.rdb, lock, m.lock(id, "idx", "available")...)
	if err == NilBatch {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	// Get Batch instance
	acquireLock(m.ctx, m.rdb, m.timeout, lock, m.lock(id, "assigned")...)
	batch, err := m.batch(id, next[0], next[1])
	releaseLock(m.ctx, m.rdb, lock, m.lock(id, "assigned")...)
	if err != nil {
		return nil, err
	}

	return batch, nil
}

func (m *RedisBatchManager) Commit(id string, itemId string) error {
	err := m.rdb.HDel(m.ctx, m.key(id, "assigned"), itemId).Err()
	if err != nil && err != redis.Nil {
		return err
	}
	return nil
}

func (m *RedisBatchManager) RollBack(id string, itemId string) error {
	// Acquire the locks
	lock := strconv.Itoa(int(time.Now().UnixMilli()))
	acquireLock(m.ctx, m.rdb, m.timeout, lock, m.lock(id, "idx", "assigned", "available")...)
	defer releaseLock(m.ctx, m.rdb, lock, m.lock(id, "idx", "assigned", "available")...)

	// Check if batch id exists
	// exists, err := m.rdb.HExists(m.ctx, m.key(id, "assigned"), itemId).Result()
	// if err != nil {
	// 	return err
	// }
	// if !exists {
	// 	return errors.New("batch manager: can not rollback an unassigned batch")
	// }
	
	// Get the batch data
	bJson, err := m.rdb.HGet(m.ctx, m.key(id, "assigned"), itemId).Result()
	if err != nil && err != redis.Nil {
		return err
	}
	var batch bm.BatchItem
	json.Unmarshal([]byte(bJson), &batch)

	// Mark this batch "failed" as its rollbacked
	// next time the client retrieve the batch
	// check the batch "HasFailed" and "Progress"
	// to continue handle the batch at where it interupted
	batch.HasFailed = true
	
	// Set the "available" with the batch data at "index - 1"
	// and future call to Next() will return this batch
	idx, err := m.rdb.Get(m.ctx, m.key(id, "idx")).Int()
	if err != nil && err != redis.Nil {
		return err
	}
	str := fmt.Sprintf("[%v, %v]", batch.Start, batch.End)
	err = m.rdb.HSet(m.ctx, m.key(id, "available"), strconv.Itoa(idx - 1), str).Err()
	if err != nil {
		return err
	}

	// Update the index value
	err = m.rdb.Decr(m.ctx, m.key(id, "idx")).Err()
	if err != nil {
		return err
	}

	return nil
}

func (m *RedisBatchManager) key(id string, k string) string {
	return fmt.Sprintf("batch_manager:%v:batch:%v:%v", m.name, id, k)
}

func (m *RedisBatchManager) lock(id string, keys ...string) []string {
	locks := make ([]string, len(keys))
	for i, k := range keys {
		locks[i] = fmt.Sprintf("batch_manager:%v:batch:%v:lock:%v", m.name, id, k)
	}
	return locks
}

func (m *RedisBatchManager) available(id string) ([2]int, error) {
	var result [2]int
	// Get index
	idx, err := m.rdb.Get(m.ctx, m.key(id, "idx")).Int()
	if err != nil && err != redis.Nil {
		return [2]int{}, err
	}
	// Nil batch
	if idx == m.len {
		return [2]int{}, NilBatch
	}
	// Increase index
	err = m.rdb.Set(m.ctx, m.key(id, "idx"), idx + 1, 0).Err()
	if err != nil {
		return [2]int{}, err
	}
	// Get next batch start, end from "available"
	exists, err := m.rdb.HExists(m.ctx, m.key(id, "available"), strconv.Itoa(idx)).Result()
	if err != nil {
		return [2]int{}, err
	}
	if !exists {
		s := m.start + idx * m.size
		e := s + m.size - 1
		if e > m.end {
			e = m.end
		}
		str := fmt.Sprintf("[%v, %v]", s, e)
		err := m.rdb.HSet(m.ctx, m.key(id, "available"), strconv.Itoa(idx), str).Err()
		if err != nil {
			return [2]int{}, err
		}
		result = [2]int {s, e}
	} else {
		val, err := m.rdb.HGet(m.ctx, m.key(id, "available"), strconv.Itoa(idx)).Result()
		if err != nil && err != redis.Nil {
			return [2]int{}, err
		}
		json.Unmarshal([]byte(val), &result)
	}

	return result, nil
}

func (m *RedisBatchManager) batch(id string, s int, e int) (*bm.BatchItem, error) {
	var batch bm.BatchItem
	itemId := fmt.Sprintf("%v-%v", s, e)
	result, err := m.rdb.HGet(m.ctx, m.key(id, "batch"), itemId).Result()
	if err != nil && err != redis.Nil {
		return nil, err
	}
	if err == redis.Nil {
		batch = bm.BatchItem{
			Id: itemId,
			Start: s,
			End: e,
			HasFailed: false,
			Progress: 0,
		}
		bJson, _ := json.Marshal(batch)
		err := m.rdb.HSet(m.ctx, m.key(id, "batch"), itemId, string(bJson)).Err()
		if err != nil {
			return nil, err
		}
	} else {
		json.Unmarshal([]byte(result), &batch)
	}

	return &batch, nil
}