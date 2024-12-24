// Implement the Source interface, this package is responsible for reading the database
// migration files and creating Migration objects. Before any read operation is requested
// the migrations files content will be buffered and wait for read requests.
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
	ErrNextMigration = "LocalFile: can not load next migration. You should call HasNext() before invoke Next()"
	ErrFileNameFormat    = func(name string) string {
		return fmt.Sprintf(
			"LocalFile: migration file format error. Filename: %v. Example of a valid name 00001_create_user_table.json",
			name,
		)
	}
)

type LocalFile struct {
	uri     string       // The migration files directory location
	err     error        // Buffer error that can be retrieve with Error()
	status  SourceStatus // Stop buffering when status is closed
	files   []string     // All migration file names
	batch	[]string	 
	fileIdx int          // Index for files iteration

	ctx       context.Context
	ctxCancel context.CancelFunc
	mu        sync.Mutex

	migrationBuffer chan migrate.Migration
	maxBufferSize   int64 // If reach max size, cease buffering and wait
	buffered        int64 // The current buffer size
}

func New(ctx context.Context, cancel context.CancelFunc, uri string) (*LocalFile, error) {
	src := LocalFile{}
	src.uri = uri
	src.ctx = ctx
	src.ctxCancel = cancel
	// the estimated for the number of migration file that will be buffered
	estFileNum := defaultMaxBufferSize / estimatedFileSize
	src.migrationBuffer = make(chan migrate.Migration, estFileNum)
	src.maxBufferSize = defaultMaxBufferSize
	src.buffered = 0
	// try to open uri
	dirEntry, err := os.ReadDir(src.uri)
	if err != nil {
		return nil, err
	}
	// list all files
	src.files = make([]string, len(dirEntry))
	for i, entry := range dirEntry {
		src.files[i] = entry.Name()
	}

	return &src, nil
}

// Migration files will be pre-loaded based on the buffer size, the default is 64KB which
// is roughly enough for 30 migration files in 2KB. You must called this function before
// calling Open() or it will not change the buffer size.
func (src *LocalFile) SetBufferSize(size int64) *LocalFile {
	if src.status != StatusOpened {
		src.maxBufferSize = size
		estFileNum := size / estimatedFileSize
		src.migrationBuffer = make(chan migrate.Migration, estFileNum)
	}
	return src
}

// Start a goroutine to buffer the migration files
//
// Parameters:
//   - batch: list of migration file names to read
func (src *LocalFile) Open(batch []string) error {
	src.batch = batch
	go src.buffer()
	src.status = StatusOpened

	return nil
}

func (src *LocalFile) List() []string {
	src.mu.Lock()
	defer src.mu.Unlock()

	return src.files
}

func (src *LocalFile) HasNext() bool {
	src.mu.Lock()
	defer src.mu.Unlock()
	if len(src.batch) == 0 || src.status == StatusClosed || src.err != nil {
		return false
	}
	if src.fileIdx < len(src.batch) {
		return true
	}

	return false
}

func (src *LocalFile) Next() (*migrate.Migration, error) {
	if !src.HasNext() {
		return nil, errors.New(ErrNextMigration)
	}

	migr := <-src.migrationBuffer

	src.mu.Lock()
	defer src.mu.Unlock()

	src.buffered -= int64(len(migr.Body))
	src.fileIdx++

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

func (src *LocalFile) buffer() {
	idx := 0
	// buffer one migration file per call
	// return false when all files have been buffered
	buffering := func() bool {
		src.mu.Lock()
		defer func() {
			src.mu.Unlock()
			if r := recover(); r != nil {
				if r, ok := r.(error); ok {
					src.setErr(r)
				}
				src.Close()
			}
		}()
		filename := src.batch[idx]
		path := filepath.Join(src.uri, filename)
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
		migr, err := createMigration(filename, body)
		if err != nil {
			panic(err)
		}
		// push migration to the buffer
		src.migrationBuffer <- *migr
		src.buffered += info.Size()
		// track index
		idx++
		
		return idx != len(src.batch)
	}
	// start buffering
	for {
		select {
		case <-src.ctx.Done():
			src.Close()
			return
		default:
			if !buffering() {
				return
			}
		}
	}
}

func createMigration(fileName string, body []byte) (*migrate.Migration, error) {
	re, _ := regexp.Compile(`^(.+?)(\.rollback)?(\.\w+)$`)
	parts := re.FindStringSubmatch(fileName)

	if len(parts) < 4 {
		return nil, errors.New(ErrFileNameFormat(fileName))
	}

	migr := migrate.Migration{
		Name: parts[1],
		Body: body,
	}

	return &migr, nil
}

func (src *LocalFile) setErr(e error) {
	src.mu.Lock()
	defer src.mu.Unlock()
	src.err = e
}
