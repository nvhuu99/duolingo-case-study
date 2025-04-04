package local

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path"
	"sync"
	"time"

	lw "duolingo/lib/log/writer"
)

type LocalWriter struct {
	Path           string
	Rotation       time.Duration
	BufferSize     int
	MaxBufferCount int

	buffered      int
	bufferedCount int
	bufferCh      chan *lw.Writable
	writeCh       chan *lw.Writable

	currentRotation  time.Time
	rotationDeadline <-chan time.Time

	ctx context.Context
	mu  sync.RWMutex
}

func NewLocalWriter(ctx context.Context, path string, bufferMb int, bufferCount int, rotation time.Duration) *LocalWriter {
	writer := &LocalWriter{
		Path:           path,
		Rotation:       rotation,
		BufferSize:     bufferMb * 1024 * 1024,
		MaxBufferCount: bufferCount,
		ctx:            ctx,
		bufferCh:       make(chan *lw.Writable, 1000),
		writeCh:        make(chan *lw.Writable, 1000),
	}

	go writer.RunWriter()

	return writer
}

func (writer *LocalWriter) Write(log *lw.Writable) {
	writer.writeCh <- log
}

func (writer *LocalWriter) RunWriter() {
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
			if writer.wouldLimitExceed(log.Content) {
				writer.flush()
			}
			writer.buffer(log)
		}
	}
}

func (writer *LocalWriter) flush() {
	writer.mu.RLock()
	timestamp := writer.currentRotation.Format("20060102150405")
	countSnapshot := writer.bufferedCount
	writer.mu.RUnlock()

	mappedByFilename := make(map[string][][]byte)
	for range countSnapshot {
		log := <-writer.bufferCh
		fullPath := path.Join(writer.Path, fmt.Sprintf("%v_%v.%v", log.Prefix, timestamp, log.Extension))
		mappedByFilename[fullPath] = append(mappedByFilename[fullPath], log.Content)
	}

	for filepath, lines := range mappedByFilename {
		writer.writeLines(filepath, lines)
	}
}

func (writer *LocalWriter) wouldLimitExceed(log []byte) bool {
	writer.mu.RLock()
	defer writer.mu.RUnlock()

	return writer.buffered+len(log) > writer.BufferSize ||
		writer.bufferedCount+1 >= writer.MaxBufferCount
}

func (writer *LocalWriter) buffer(log *lw.Writable) {
	writer.bufferCh <- log

	writer.mu.Lock()
	defer writer.mu.Unlock()

	writer.buffered += len(log.Content)
	writer.bufferedCount++
}

func (writer *LocalWriter) rotate() {
	writer.mu.Lock()
	defer writer.mu.Unlock()

	writer.rotationDeadline = time.After(writer.Rotation)
	writer.currentRotation = time.Now()
}

func (writer *LocalWriter) writeLines(filepath string, lines [][]byte) {
	os.MkdirAll(path.Dir(filepath), 0755)
	file, _ := os.OpenFile(filepath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	defer file.Close()

	wr := bufio.NewWriter(file)
	defer wr.Flush()

	for _, line := range lines {
		wr.Write(append(line, '\n'))
	}

	writer.mu.Lock()
	defer writer.mu.Unlock()

	totalBytes := 0
	for _, line := range lines {
		totalBytes += len(line)
	}
	writer.buffered -= totalBytes
	writer.bufferedCount -= len(lines)
}
