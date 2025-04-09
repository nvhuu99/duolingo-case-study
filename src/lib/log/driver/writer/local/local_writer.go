package local

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path"
	"strings"
	"sync"
	"time"

	lw "duolingo/lib/log/writer"
)

type LocalWriter struct {
	Path           string
	Rotation       time.Duration
	FlushInterval  time.Duration
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

func NewLocalWriter(ctx context.Context, path string) *LocalWriter {
	writer := &LocalWriter{
		Path:           path,
		Rotation:       time.Hour,
		BufferSize:     1,
		MaxBufferCount: 1000,
		ctx:            ctx,
		bufferCh:       make(chan *lw.Writable, 1000),
		writeCh:        make(chan *lw.Writable, 1000),
	}

	go writer.RunWriter()

	return writer
}

func (writer *LocalWriter) WithBuffering(sizeMb int, maxCount int) lw.LogWriter {
	writer.BufferSize = sizeMb
	writer.MaxBufferCount = maxCount
	writer.bufferCh = make(chan *lw.Writable, 100+writer.MaxBufferCount)
	writer.writeCh = make(chan *lw.Writable, 100+writer.MaxBufferCount)
	return writer
}

func (writer *LocalWriter) WithRotation(interval time.Duration) lw.LogWriter {
	writer.Rotation = interval
	return writer
}

func (writer *LocalWriter) WithFlushInterval(interval time.Duration) lw.LogWriter {
	writer.FlushInterval = interval
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
			writer.buffer(log)
			if writer.hasLimitExceeded() {
				writer.flush()
			}
		}
	}
}

func (writer *LocalWriter) flush() {
	writer.mu.RLock()
	countSnapshot := writer.bufferedCount
	datetime := writer.currentRotation.Format("20060102150405")
	year := fmt.Sprintf("%d", writer.currentRotation.Year())
	month := fmt.Sprintf("%02d", int(writer.currentRotation.Month()))
	day := fmt.Sprintf("%02d", writer.currentRotation.Day())
	writer.mu.RUnlock()

	mappedByFilename := make(map[string][][]byte)
	for range countSnapshot {
		log := <-writer.bufferCh
		fileName := fmt.Sprintf("%v_%v.%v", log.Prefix, datetime, log.Extension)
		fullPath := path.Join(
			writer.Path,
			path.Join(strings.Split(log.Namespace, "/")...),
			year,
			month,
			day,
			fileName,
		)
		mappedByFilename[fullPath] = append(mappedByFilename[fullPath], log.Content)
	}

	for filepath, lines := range mappedByFilename {
		writer.writeLines(filepath, lines)
	}
}

func (writer *LocalWriter) hasLimitExceeded() bool {
	writer.mu.RLock()
	defer writer.mu.RUnlock()

	return writer.buffered > writer.BufferSize ||
		writer.bufferedCount >= writer.MaxBufferCount
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
