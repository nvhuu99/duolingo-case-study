package writer

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type LogWriter struct {
	Rotation       time.Duration
	FlushInterval  time.Duration
	FlushGraceTO   time.Duration
	BufferSize     int
	MaxBufferCount int

	outputs []LogOutput

	buffered      int
	bufferedCount int
	bufferCh      chan *Writable
	writeCh       chan *Writable

	currentRotation  time.Time
	rotationDeadline <-chan time.Time

	ctx context.Context
	mu  sync.RWMutex
}

func NewLogWriter(ctx context.Context) *LogWriter {
	writer := &LogWriter{
		Rotation:       time.Hour,
		BufferSize:     1,
		MaxBufferCount: 1000,
		FlushInterval:  10 * time.Second,
		FlushGraceTO:   300 * time.Millisecond,
		ctx:            ctx,
		bufferCh:       make(chan *Writable, 1000),
		writeCh:        make(chan *Writable, 1000),
	}

	go writer.runWriter()

	return writer
}

func (writer *LogWriter) WithBuffering(sizeMb int, maxCount int) *LogWriter {
	writer.BufferSize = sizeMb
	writer.MaxBufferCount = maxCount
	writer.bufferCh = make(chan *Writable, 100+writer.MaxBufferCount)
	writer.writeCh = make(chan *Writable, 100+writer.MaxBufferCount)
	return writer
}

func (writer *LogWriter) WithRotation(interval time.Duration) *LogWriter {
	writer.Rotation = interval
	return writer
}

func (writer *LogWriter) WithFlushInterval(interval time.Duration, grace time.Duration) *LogWriter {
	writer.FlushGraceTO = grace
	writer.FlushInterval = interval
	return writer
}

func (writer *LogWriter) AddLogOutput(output LogOutput) *LogWriter {
	writer.outputs = append(writer.outputs, output)
	return writer
}

func (writer *LogWriter) Write(log *Writable) {
	writer.writeCh <- log
}

func (writer *LogWriter) runWriter() {
	ctx, cancel := context.WithCancel(writer.ctx)
	defer cancel()

	writer.rotate()

	for {
		select {
		case <-ctx.Done():
			writer.flush()
			return
		case <-writer.rotationDeadline:
			writer.flush()
			writer.rotate()
		case log := <-writer.writeCh:
			writer.buffer(log)
			if writer.hasLimitExceeded() {
				writer.flush()
			}
		}
	}
}

func (writer *LogWriter) flush() {
	writer.mu.RLock()
	countSnapshot := writer.bufferedCount
	writer.mu.RUnlock()

	items := make([]*Writable, countSnapshot)
	for i := range countSnapshot {
		writable := <-writer.bufferCh
		writable.Rotation = writer.currentRotation.Format("20060102150405")
		items[i] = writable
	}

	writer.writeAll(items)

	time.Sleep(writer.FlushGraceTO)
}

func (writer *LogWriter) hasLimitExceeded() bool {
	writer.mu.RLock()
	defer writer.mu.RUnlock()

	return writer.buffered > writer.BufferSize ||
		writer.bufferedCount >= writer.MaxBufferCount
}

func (writer *LogWriter) buffer(log *Writable) {
	writer.bufferCh <- log

	writer.mu.Lock()
	defer writer.mu.Unlock()

	writer.buffered += len(log.Content)
	writer.bufferedCount++
}

func (writer *LogWriter) rotate() {
	writer.mu.Lock()
	defer writer.mu.Unlock()
	// Calculate next rotation interval in seconds
	now := time.Now()
	interval := int(writer.Rotation.Seconds())
	dayPassedSeconds := now.Hour()*3600 + now.Minute()*60 + now.Second()
	alignedSeconds := (dayPassedSeconds / interval) * interval
	// Create a new time aligned to the rotation interval
	writer.currentRotation = time.
		Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).
		Add(time.Duration(alignedSeconds) * time.Second)
	// Reset rotation deadline to flush the buffer
	writer.rotationDeadline = time.After(writer.Rotation)
}

func (writer *LogWriter) writeAll(items []*Writable) {
	for _, opt := range writer.outputs {
		if err := opt.Flush(items); err != nil {
			fmt.Println(err)
		}
	}

	writer.mu.Lock()
	defer writer.mu.Unlock()

	for _, line := range items {
		writer.buffered -= len(line.Content)
		writer.bufferedCount--
	}
}
