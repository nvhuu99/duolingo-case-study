package config_reader

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

type LocalFile struct {
	dir        string
	acceptExts []string
	files      []string
}

func NewLocalFile(dir string) *LocalFile {
	return &LocalFile{
		dir: dir,
	}
}

func (src *LocalFile) AcceptExtentions(ext ...string) *LocalFile {
	for i := range ext {
		if !slices.Contains(src.acceptExts, ext[i]) {
			src.acceptExts = append(src.acceptExts, ext[i])
		}
	}
	return src
}

func (src *LocalFile) Load(filename string) ([][]byte, error) {
	if src.dir == "" {
		panic(fmt.Errorf(ErrSourceFailure, "LocalFile", "source dir is not set"))
	}

	if len(src.files) == 0 {
		src.files = src.listFiles()
	}

	var result [][]byte

	for _, path := range src.files {
		info, err := os.Stat(path)
		if err != nil || !strings.HasPrefix(info.Name(), filename) {
			continue
		}
		content, err := os.ReadFile(path)
		if err != nil {
			panic(fmt.Errorf(ErrSourceFailure, src, err))
		}
		result = append(result, content)
	}

	return result, nil
}

func (src *LocalFile) listFiles() []string {
	found := []string{}
	err := filepath.Walk(src.dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		name := info.Name()
		fullPath, _ := filepath.Abs(path)
		for _, ext := range src.acceptExts {
			if strings.HasSuffix(name, ext) {
				found = append(found, fullPath)
				break
			}
		}
		return nil
	})
	if err != nil {
		panic(fmt.Errorf(ErrSourceFailure, src, err))
	}
	return found
}

// Implement Stringer interface for error messages
func (src *LocalFile) String() string {
	if src.dir == "" {
		return "LocalFile"
	}
	return fmt.Sprintf("LocalFile(%v)", src.dir)
}
