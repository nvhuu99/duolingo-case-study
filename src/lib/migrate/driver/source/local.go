// Implement the Source interface, this package is response for reading the database
// migration files and creating Migration objects. Before any read operation is requested
// the migrations files content will be buffered and wait for read requests.
//
// Example usage:
//
// 	ctx, cancel := context.WithCancel(context.Background())
// 	// set up
// 	src := local.New(ctx, cancel, "database/migrations")
// 		.SetBufferSize(128 * 1024)
// 		.SetOperationTimeOut(30 * time.Second)
// 		.SetMigrateType(local.MigrateTypeUp)
// 	err := src.Open()
// 	if err != nil {
// 		panic(err)
// 	}
// 	// some error need to be catched (when buffering)
// 	go func() {
// 		<-ctx.Done()
// 		if src.HasError() {
// 			err := src.Error()
// 		}
// 	}()
// 	// load all migrations
// 	for src.HasNext() {
// 		m, err := src.Next()
// 		if err != nil {
// 			cancel()
// 			panic(err.Error())
// 		}
// 		// use the migrations ...
// 	}
package local

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sync"
	"time"

	migrate "duolingo/lib/migrate"
)

const (
	readWait             = 5 * time.Second
	defaultTimeOut = 30 * time.Second
	defaultMaxBufferSize = 64 * 1024
	estimatedFileSize    = 2 * 1024
)

const (
	NilVersion          string        = "-1"
	MigrateUp           MigrateType   = "up"
	MigrateRollback     MigrateType   = "rollback"
	MigrateStatusOpened MigrateStatus = "opened"
	MigrateStatusClosed MigrateStatus = "closed"
)

var (
	ErrEmptyMigration = "nothing to migrate"
	ErrNextMigration  = "can not load next migration for current version"
	ErrFileNameFormat = func(name string) string {
		return fmt.Sprintf(
			"migration file format error. Example of a valid name 00001_create_user_table.json\nfilename:%v",
			name,
		)
	}
	ErrOperationTimeOut = func(op string) string {
		return fmt.Sprintf(
			"operation did not complete before timeout\noperation:%v",
			op,
		)
	}

)

type MigrateType string

type MigrateStatus string

type LocalFile struct {
	uri             string
	ver             string
	orderedVersions []string
	files           map[string]string
	migrateTye      MigrateType
	verIdx          int
	err             error

	ctx             context.Context
	ctxCancel       context.CancelFunc
	mu              sync.Mutex
	status          MigrateStatus
	migrationBuffer chan migrate.Migration
	maxBufferSize   int64
	buffered        int64
	timeOut			time.Duration
}

// All files inside the directory in the provided uri must be in the correct format
func New(ctx context.Context, cancel context.CancelFunc, uri string, ver string) *LocalFile {
	src := LocalFile{}
	src.uri = uri
	src.ver = ver
	src.migrateTye = MigrateUp
	src.ctx = ctx
	src.ctxCancel = cancel
	// the estimated for the number of migration file that will be buffered
	estFileNum := defaultMaxBufferSize / estimatedFileSize
	src.migrationBuffer = make(chan migrate.Migration, estFileNum)
	src.maxBufferSize = defaultMaxBufferSize
	src.buffered = 0
	src.timeOut = defaultTimeOut
	return &src
}

// Migration files will be pre-loaded based on the buffer size, the default is 64KB which
// is roughly enough for 30 migration files in 2KB. You must called this function before
// calling Open() or it will not change the buffer size.
func (src *LocalFile) SetBufferSize(size int64) *LocalFile {
	if len(src.orderedVersions) == 0 {
		src.maxBufferSize = size
		estFileNum := size / estimatedFileSize
		src.migrationBuffer = make(chan migrate.Migration, estFileNum)
	}
	return src
}

// You must called this function before calling Open() or the buffer will behave uncorrectly.
func (src *LocalFile) SetMigrateType(direction MigrateType) *LocalFile {
	if len(src.orderedVersions) == 0 {
		src.migrateTye = direction
	}
	return src
}

// You must called this function before calling Open() or the buffer will behave uncorrectly.
func (src *LocalFile) SetOperationTimeOut(duration time.Duration) *LocalFile {
	src.timeOut = duration
	return src
}

// Read all migration files in the location according to the uri. Then start a goroutine to buffer the migration files
func (src *LocalFile) Open() error {
	done := make(chan any)
	src.startTimeOut("Open()", done)
	// try to open uri
	dirEntry, err := os.ReadDir(src.uri)
	if err != nil {
		return err
	}
	if len(dirEntry) == 0 {
		return errors.New(ErrEmptyMigration)
	}
	// list all files, store versions in ascending order
	src.orderedVersions = make([]string, len(dirEntry))
	src.files = make(map[string]string, len(dirEntry))
	for i, entry := range dirEntry {
		re, _ := regexp.Compile(`([^_]+)_(.*)(\.[a-z]+$)`)
		parts := re.FindStringSubmatch(entry.Name())
		if len(parts) < 4 {
			return errors.New(ErrFileNameFormat(entry.Name()))
		}
		src.orderedVersions[i] = parts[1]
		src.files[parts[1]] = entry.Name()
		if src.ver != NilVersion && parts[1] == src.ver {
			src.verIdx = i
		}
	}
	// start buffering
	go src.buffer()
	// set status and finish
	src.status = MigrateStatusOpened
	done <- struct{}{}

	return nil
}

func (src *LocalFile) startTimeOut(name string, done chan any) {
	var check bool
	var shouldSetErr bool
	var mu sync.Mutex 
	// extend timeout slightly to ensure "done" is processed first
	timeOut, cancel := context.WithTimeout(src.ctx, src.timeOut + time.Second)
	go func() {
		defer cancel() 
		defer close(done)
		for {
			select {
			case <-done:
				mu.Lock()
				check = true
				mu.Unlock()
				return
			case <-src.ctx.Done():
				src.Close()
				return
			case <-timeOut.Done():
				mu.Lock()
				if !check {
					shouldSetErr = true
				}
				mu.Unlock()
				if shouldSetErr {
					src.setErr(errors.New(ErrOperationTimeOut(name)))
					src.terminate()
				}
				return
			}
		}
	}()
}

func (src *LocalFile) buffer() {
	// wrap the block for easy mutex control
	buffering := func() {
		src.mu.Lock()
		defer func() {
			src.mu.Unlock()
			// handle panic while loading the migrations
			if r := recover(); r != nil {
				src.setErr(r.(error))
				src.terminate()
			}
		}()
		idx := src.verIdx
		v := src.orderedVersions[idx]
		f := src.files[v]
		path := filepath.Join(src.uri, f)
		// only continue if satisfy the max size
		info, err := os.Stat(path)
		if err != nil {
			panic(err)
		}
		if src.buffered+info.Size() > src.maxBufferSize {
			return
		}
		// read the file and create a Migration
		body, err := os.ReadFile(path)
		if err != nil {
			panic(err)
		}
		migr := createMigration(f)
		migr.Body = body
		// push migration to the buffer
		src.migrationBuffer <- *migr
		// keep tracking on the buffer size and buffer index
		if src.migrateTye == MigrateUp {
			src.verIdx++
		} else {
			src.verIdx--
		}
		src.buffered += info.Size()
	}
	// start buffering
	for {
		select {
		case <-src.ctx.Done():
			return
		default:
			if !src.HasNext() {
				return
			}
			done := make(chan any)
			src.startTimeOut("internal: buffer()", done)
			buffering()
			done <- struct{}{}
		}
	}
}

// Init the Migration object by extracting version, and name from the file
func createMigration(filename string) *migrate.Migration {
	re, _ := regexp.Compile(`([^_]+)_(.*)(\.[a-z]+$)`)
	parts := re.FindStringSubmatch(filename)
	migr := migrate.Migration{
		Version: parts[1],
		Name:    parts[2],
	}
	return &migr
}

// Return the map of migration files by it's versions.
func (src *LocalFile) List() map[string]string {
	src.mu.Lock()
	defer src.mu.Unlock()
	copy := make(map[string]string, len(src.files))
	for k, v := range src.files {
		copy[k] = v
	}
	return copy
}

// The total migration files
func (src *LocalFile) Count() int {
	src.mu.Lock()
	defer src.mu.Unlock()
	return len(src.orderedVersions)
}

// Determie whether the database can migrate up. Return false when at the lastest version.
func (src *LocalFile) HasNext() bool {
	src.mu.Lock()
	defer src.mu.Unlock()
	if len(src.orderedVersions) == 0 || src.status == MigrateStatusClosed || src.err != nil {
		return false
	}
	lastVer := len(src.orderedVersions) - 1
	if src.verIdx < lastVer {
		return true
	}
	return false
}

// Determie whether the database can be rollbacked. Return false when at the first version.
func (src *LocalFile) HasPrev() bool {
	src.mu.Lock()
	defer src.mu.Unlock()
	if len(src.orderedVersions) == 0 || src.status == MigrateStatusClosed || src.err != nil {
		return false
	}
	if src.verIdx > 0 {
		return true
	}
	return false
}

// Return a Migration to migrate up the database
func (src *LocalFile) Next() (*migrate.Migration, error) {
	if !src.HasNext() {
		return nil, errors.New(ErrNextMigration)
	}
	migrate := <-src.migrationBuffer
	src.mu.Lock()
	src.buffered -= int64(len(migrate.Body))
	src.mu.Unlock()
	go src.buffer()

	return &migrate, nil
}

// Return a Migration to rollback the database
// TODO
func (src *LocalFile) Prev() (*migrate.Migration, error) {
	return nil, nil
}

func (src *LocalFile) Close() {
	src.mu.Lock()
	defer src.mu.Unlock()
	if src.status != MigrateStatusClosed {
		src.status = MigrateStatusClosed
		close(src.migrationBuffer)
	}
}

// Call context cancel when a unhandled error occurs and needs to be handled
// by the parent goroutine. The error is expected to occur when trying to buffer.
func (src *LocalFile) terminate() {
	src.ctxCancel()
	src.Close()
}

// Determine if there any unhanled error
func (src *LocalFile) HasError() bool {
	return src.err != nil
}

// Return the unhandled error
func (src *LocalFile) Error() error {
	return src.err
}

// Set unhandled error
func (src *LocalFile) setErr(e error) {
	src.mu.Lock()
	defer src.mu.Unlock()
	src.err = e
}
