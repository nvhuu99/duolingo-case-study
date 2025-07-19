package config_reader

import (
	"os"
	"sync/atomic"
)

type LocalFile struct {
	filepath string
	cache    []byte
	loaded   atomic.Bool
}

func NewLocalFile(filepath string) *LocalFile {
	return &LocalFile{
		filepath: filepath,
	}
}

func (src *LocalFile) Load() ([]byte, error) {
	if src.loaded.Load() {
		return src.cache, nil
	}
	if _, err := os.Stat(src.filepath); err != nil {
		return nil, err
	}
	content, err := os.ReadFile(src.filepath)
	if err == nil {
		src.cache = content
		src.loaded.Store(true)
	}
	return content, err
}
