// Implement the Source interface, this package is responsible for reading the database
// migration files and creating Migration objects. Before any read operation is requested
// the migrations files content will be buffered and wait for read requests.
//
// Example usage:
//
//	ctx, cancel := context.WithCancel(context.Background())
//	// set up
//	src := local.New(ctx, cancel, "database/migrations")
//		.SetBufferSize(128 * 1024)
//		.SetMigrateType(local.MigrateTypeUp)
//	err := src.Open(migrate.NilVersion)
//	if err != nil {
//		panic(err)
//	}
//	// some error need to be catched (when buffering)
//	go func() {
//		<-ctx.Done()
//		if src.HasError() {
//			err := src.Error()
//		}
//	}()
//	// load all migrations
//	for src.HasNext() {
//		m, err := src.Next()
//		if err != nil {
//			cancel()
//			panic(err.Error())
//		}
//		// use the migrations ...
//	}
package local

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sync"

	migrate "duolingo/lib/migrate"
)

const (
	defaultMaxBufferSize = 64 * 1024
	estimatedFileSize    = 2 * 1024
)

type SourceStatus string

const (
	StatusOpened SourceStatus = "opened"
	StatusClosed SourceStatus = "closed"
)

var (
	ErrEmptyMigration    = "LocalFile: nothing to migrate"
	ErrNextMigration     = "LocalFile: can not load next migration for current version. You should call HasNext() before invoke Next()"
	ErrPreviousMigration = "LocalFile: can not load previous migration for current version. You should call HasPrev() before invoke Prev()"
	ErrFileNameFormat    = func(name string) string {
		return fmt.Sprintf(
			"LocalFile: migration file format error. Filename: %v. Example of a valid name 00001_create_user_table.json",
			name,
		)
	}
)

type LocalFile struct {
	uri             string				// The migration files directory location  
	orderedVersions []string			// Versions extracted from the migration files
	files           map[string]string   // The migration file paths mapped by versions
	migrateTye      migrate.MigrateType // The migration direction (up or rollback)
	verIdx          int                 // The index to iterate over the migrations
	err             error               // Buffer error that can be retrieve with Error()

	ctx       context.Context
	ctxCancel context.CancelFunc
	mu        sync.Mutex

	migrationBuffer chan migrate.Migration		
	maxBufferSize   int64						// If reach max size, cease buffering and wait
	buffered        int64						// The current buffer size
	status          SourceStatus				// Stop buffering when status is closed
}

func New(ctx context.Context, cancel context.CancelFunc, uri string) *LocalFile {
	src := LocalFile{}
	src.uri = uri
	src.migrateTye = migrate.MigrateUp
	src.ctx = ctx
	src.ctxCancel = cancel
	// the estimated for the number of migration file that will be buffered
	estFileNum := defaultMaxBufferSize / estimatedFileSize
	src.migrationBuffer = make(chan migrate.Migration, estFileNum)
	src.maxBufferSize = defaultMaxBufferSize
	src.buffered = 0
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

// Set migration direction (up or rollback).
// You must called this function before calling Open() or the buffer will behave uncorrectly.
func (src *LocalFile) SetMigrateType(direction migrate.MigrateType) *LocalFile {
	if len(src.orderedVersions) == 0 {
		src.migrateTye = direction
	}
	return src
}

// Read all migration files in the location according to the uri. 
// Then start a goroutine to buffer the migration files
//
// Parameters:
//   - ver: The current database version retrieved from the Database driver
func (src *LocalFile) Open(ver string) error {
	// try to open uri
	dirEntry, err := os.ReadDir(src.uri)
	if err != nil {
		return err
	}
	// list all files, store versions in ascending order
	src.orderedVersions = make([]string, len(dirEntry))
	src.files = make(map[string]string, len(dirEntry))
	for i, entry := range dirEntry {
		re, _ := regexp.Compile(`([0-9]+)_(.*)(\.[a-z]+$)`)
		parts := re.FindStringSubmatch(entry.Name())
		if len(parts) < 4 {
			return errors.New(ErrFileNameFormat(entry.Name()))
		}
		src.orderedVersions[i] = parts[1]
		src.files[parts[1]] = entry.Name()
	}
	// set version index
	src.setVersionIndex(ver)
	if (src.migrateTye == migrate.MigrateUp && src.verIdx == len(src.orderedVersions)) ||
		(src.migrateTye == migrate.MigrateRollback && src.verIdx < 0) {
		return errors.New(ErrEmptyMigration)
	}
	// start buffering
	go src.buffer()
	// set status and finish
	src.status = StatusOpened

	return nil
}

func (src *LocalFile) setVersionIndex(version string) {
	for idx, ver := range src.orderedVersions {
		if version == ver {
			if src.migrateTye == migrate.MigrateUp {
				src.verIdx = idx + 1
			} else {
				src.verIdx = idx - 1
			}
			return
		}
	}
}

func (src *LocalFile) buffer() {
	idx := src.verIdx
	// buffer one migration file per call
	// return false when all files have been buffered
	buffering := func() bool {
		src.mu.Lock()
		defer func() {
			src.mu.Unlock()
			if r := recover(); r != nil {
				src.setErr(r.(error))
				src.Close()
			}
		}()
		v := src.orderedVersions[idx]
		f := src.files[v]
		path := filepath.Join(src.uri, f)
		// only continue if satisfy the max size
		info, err := os.Stat(path)
		if err != nil {
			panic(err)
		}
		if src.buffered+info.Size() > src.maxBufferSize {
			return true
		}
		// read the file and create a Migration
		body, err := os.ReadFile(path)
		if err != nil {
			panic(err)
		}
		migr := createMigration(f, body)
		// push migration to the buffer
		src.migrationBuffer <- *migr
		src.buffered += info.Size()
		// track index
		if src.migrateTye == migrate.MigrateUp {
			idx++
		} else {
			idx--
		}
		if idx < 0 || idx == len(src.orderedVersions) {
			return false
		}
		
		return true
	}
	// start buffering
	for {
		select {
		case <-src.ctx.Done():
			src.Close()
			return
		default:
			if ! buffering() {
				return
			}
		}
	}
}

func createMigration(filename string, body []byte) *migrate.Migration {
	re, _ := regexp.Compile(`([^_]+)_(.*)(\.[a-z]+$)`)
	parts := re.FindStringSubmatch(filename)
	migr := migrate.Migration{
		Version: parts[1],
		Name:    filename,
		Body: 	 body,
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
	if len(src.orderedVersions) == 0 || src.status == StatusClosed || src.err != nil {
		return false
	}
	if src.verIdx < len(src.orderedVersions) {
		return true
	}

	return false
}

// Determie whether the database can be rollbacked. Return false when at the first version.
func (src *LocalFile) HasPrev() bool {
	src.mu.Lock()
	defer src.mu.Unlock()
	if len(src.orderedVersions) == 0 || src.status == StatusClosed || src.err != nil {
		return false
	}
	if src.verIdx >= 0 {
		return true
	}
	return false
}

// Return a Migration to migrate up the database
func (src *LocalFile) Next() (*migrate.Migration, error) {
	if !src.HasNext() {
		return nil, errors.New(ErrNextMigration)
	}

	migr := <-src.migrationBuffer

	src.mu.Lock()
	defer src.mu.Unlock()

	src.buffered -= int64(len(migr.Body))
	src.verIdx++

	return &migr, nil
}

// Return a Migration to rollback the database
func (src *LocalFile) Prev() (*migrate.Migration, error) {
	if !src.HasPrev() {
		return nil, errors.New(ErrPreviousMigration)
	}

	migr := <-src.migrationBuffer

	src.mu.Lock()
	defer src.mu.Unlock()

	src.buffered -= int64(len(migr.Body))
	src.verIdx--

	return &migr, nil
}

func (src *LocalFile) Close() {
	src.mu.Lock()
	defer src.mu.Unlock()
	if src.status != StatusClosed {
		src.status = StatusClosed
		close(src.migrationBuffer)
	}
}

// Determine if there any unhanled error
func (src *LocalFile) HasError() bool {
	return src.err != nil
}

// Return the unhandled error
func (src *LocalFile) Error() error {
	return src.err
}

func (src *LocalFile) setErr(e error) {
	src.mu.Lock()
	defer src.mu.Unlock()
	src.err = e
}
