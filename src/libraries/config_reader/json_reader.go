package config_reader

import (
	"fmt"
	"path"

	"github.com/tidwall/gjson"
)

type JsonConfigReader struct {
	sources      map[string]Source
	localFileDir string
}

func NewJsonConfigReader() *JsonConfigReader {
	return &JsonConfigReader{
		sources: make(map[string]Source),
	}
}

func (reader *JsonConfigReader) SetLocalFileDir(dir string) *JsonConfigReader {
	reader.localFileDir = dir
	return reader
}

func (reader *JsonConfigReader) AddLocalFile(name string, filepath string) *JsonConfigReader {
	if _, exists := reader.sources[name]; !exists {
		fullpath := path.Join(reader.localFileDir, filepath)
		reader.sources[name] = NewLocalFile(fullpath)
	}
	return reader
}

func (reader *JsonConfigReader) Get(source string, pattern string) string {
	return reader.get(source, pattern).String()
}

func (reader *JsonConfigReader) GetInt(source string, pattern string) int {
	return int(reader.get(source, pattern).Int())
}

func (reader *JsonConfigReader) GetInt64(source string, pattern string) int64 {
	return reader.get(source, pattern).Int()
}

func (reader *JsonConfigReader) GetArr(source string, pattern string) []string {
	result := []string{}
	data := reader.get(source, pattern).Array()
	for r := range data {
		if data[r].Exists() {
			result = append(result, data[r].String())
		}
	}
	return result
}

func (reader *JsonConfigReader) GetIntArr(source string, pattern string) []int {
	result := []int{}
	data := reader.get(source, pattern).Array()
	for r := range data {
		if data[r].Exists() {
			result = append(result, int(data[r].Int()))
		}
	}
	return result
}

func (reader *JsonConfigReader) GetInt64Arr(source string, pattern string) []int64 {
	result := []int64{}
	data := reader.get(source, pattern).Array()
	for r := range data {
		if data[r].Exists() {
			result = append(result, data[r].Int())
		}
	}
	return result
}

func (reader *JsonConfigReader) get(source string, pattern string) gjson.Result {
	src, exists := reader.sources[source]
	if !exists {
		panic(fmt.Errorf(ErrSourceNotRegistered, source))
	}
	raw, err := src.Load()
	if err != nil {
		panic(fmt.Errorf(ErrSourceFailure, source, err))
	}
	data := gjson.ParseBytes(raw).Get(pattern)
	if !data.Exists() {
		panic(fmt.Errorf(ErrConfigNotFound, pattern, source))
	}
	return data
}
