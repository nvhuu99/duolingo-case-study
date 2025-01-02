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
	defaultMaxBufferSize = 64 * 1024 // Default buffer size for migration files (64 KB).
	estimatedFileSize    = 2 * 1024  // Estimated average size of a migration file (2 KB).
)

type sourceStatus string

const (
	statusOpened sourceStatus = "opened" // Source is actively buffering migration files.
	statusClosed sourceStatus = "closed" // Source has been closed; buffering has stopped.
)

var (
	errNextMigration  = "LocalFile: Cannot load the next migration. Ensure HasNext() is called before invoking Next()."
	errFileNameFormat = func(name string) string {
		return fmt.Sprintf(
			"LocalFile: Invalid migration file name format. Filename: %v. Expected format: 00001_create_user_table.json.",
			name,
		)
	}
)

// LocalFile implements the Source interface, managing database migration files from a local directory.
// Files are pre-buffered for efficient processing based on a configurable buffer size.
type LocalFile struct {
	uri            string       // Directory location for migration files.
	err            error        // Tracks the most recent error.
	status         sourceStatus // Tracks the current status of the source (opened/closed).
	files          []string     // List of all migration file names in the directory.
	batch          []string     // Current batch of migration files to process.
	fileIdx        int          // Index for iterating through the batch.

	ctx            context.Context       // Context for managing goroutine lifecycle.
	ctxCancel      context.CancelFunc    // Function to cancel the context.
	mu             sync.Mutex            // Mutex for safe concurrent access.

	migrationBuffer chan migrate.Migration // Channel for buffered Migration objects.
	maxBufferSize   int64                  // Maximum size of the buffer.
	buffered        int64                  // Current size of the buffered data.
}


// New creates and initializes a new LocalFile instance.
func New(ctx context.Context, cancel context.CancelFunc) *LocalFile {
	src := LocalFile{}
	src.ctx = ctx
	src.ctxCancel = cancel

	// Estimate the initial buffer capacity based on default buffer size and file size.
	estFileNum := defaultMaxBufferSize / estimatedFileSize
	src.migrationBuffer = make(chan migrate.Migration, estFileNum)
	src.maxBufferSize = defaultMaxBufferSize
	src.buffered = 0

	return &src
}

// UseUri sets the URI (directory path) for migration files and lists all available files.
func (src *LocalFile) UseUri(uri string) error {
	src.uri = uri
	dirEntry, err := os.ReadDir(src.uri)
	if err != nil {
		return err
	}

	src.files = make([]string, len(dirEntry))
	for i, entry := range dirEntry {
		src.files[i] = entry.Name()
	}

	return nil
}

// SetBufferSize adjusts the buffer size for preloading migration files.
// This must be called before Open() to take effect.
func (src *LocalFile) SetBufferSize(size int64) *LocalFile {
	if src.status != statusOpened {
		src.maxBufferSize = size
		estFileNum := size / estimatedFileSize
		src.migrationBuffer = make(chan migrate.Migration, estFileNum)
	}
	return src
}

// Open starts buffering migration files from the given batch in a separate goroutine.
// A batch is a ordered list of files to be buffered
func (src *LocalFile) Open(batch []string) error {
	src.batch = batch
	go src.buffer()
	src.status = statusOpened
	return nil
}

// List returns the names of all available migration files.
func (src *LocalFile) List() []string {
	src.mu.Lock()
	defer src.mu.Unlock()
	return src.files
}

// HasNext checks if there are more migration files to process in the batch.
func (src *LocalFile) HasNext() bool {
	src.mu.Lock()
	defer src.mu.Unlock()

	if len(src.batch) == 0 || src.status == statusClosed || src.err != nil {
		return false
	}
	return src.fileIdx < len(src.batch)
}

// Next retrieves the next migration file from the buffer
func (src *LocalFile) Next() (*migrate.Migration, error) {
	if !src.HasNext() {
		return nil, errors.New(errNextMigration)
	}

	migr := <-src.migrationBuffer
	src.mu.Lock()
	defer src.mu.Unlock()

	src.buffered -= int64(len(migr.Body))
	src.fileIdx++
	
	return &migr, nil
}

// Close stops the buffering operation and closes the migration buffer.
func (src *LocalFile) Close() {
	src.mu.Lock()
	defer src.mu.Unlock()
	if src.status != statusClosed {
		src.status = statusClosed
		close(src.migrationBuffer)
		src.ctxCancel()
	}
}

// HasError checks if there is an unhandled error in the source.
func (src *LocalFile) HasError() bool {
	return src.err != nil
}

// Error returns the most recent unhandled error.
func (src *LocalFile) Error() error {
	return src.err
}

// buffer is a goroutine function for preloading migration files into the buffer.
func (src *LocalFile) buffer() {
	idx := 0

	// buffering loads one file per call, handling errors and context cancellation
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
		return nil, errors.New(errFileNameFormat(fileName))
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
